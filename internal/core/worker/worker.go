package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/core/renamer"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
	"gorm.io/gorm"
)

// Job 代表一个待执行的转存任务
type Job struct {
	Task *db.Task
}

// Manager 负责管理 Worker 池 and 任务分发
type Manager struct {
	workers  int
	jobQueue chan Job
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	db       *gorm.DB
}

func NewManager(numWorkers int, dbInst *gorm.DB) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		workers:  numWorkers,
		jobQueue: make(chan Job, 100),
		ctx:      ctx,
		cancel:   cancel,
		db:       dbInst,
	}
}

// Start 启动所有 Worker
func (m *Manager) Start() {
	for i := 0; i < m.workers; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}
}

// Stop 停止所有 Worker
func (m *Manager) Stop() {
	m.cancel()
	close(m.jobQueue)
	m.wg.Wait()
}

// Submit 提交一个任务到队列
func (m *Manager) Submit(job Job) {
	m.jobQueue <- job
}

func (m *Manager) worker(id int) {
	defer m.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case <-m.ctx.Done():
			log.Printf("Worker %d stopping", id)
			return
		case job, ok := <-m.jobQueue:
			if !ok {
				return
			}
			m.execute(job.Task)
		}
	}
}

func (m *Manager) updateProgress(task *db.Task, percent int, stage, message string) {
	task.Percent = percent
	task.Stage = stage
	task.Message = message
	m.db.Model(task).Updates(map[string]interface{}{
		"percent": percent,
		"stage":   stage,
		"message": message,
	})
	log.Printf("[PROGRESS:%d:%d:%s:%s]", task.ID, percent, stage, message)
	utils.BroadcastTaskUpdate(task)
}

func (m *Manager) execute(task *db.Task) {
	log.Printf("Executing task: %s (ID: %d)", task.Name, task.ID)
	m.updateProgress(task, 5, "Started", "任务已进入执行队列")

	// 1. 更新任务状态为 running
	m.db.Model(task).Update("status", "running")

	driver := core.GetDriver(&task.Account)
	if driver == nil {
		m.finishTask(task, "failed", "Driver not found")
		return
	}

	// 2. 执行转存逻辑
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Minute)
	defer cancel()

	// 2.1 获取分享全量文件
	m.updateProgress(task, 15, "Parsing", "正在解析分享链接...")
	files, err := driver.ParseShare(ctx, task.ShareURL, task.ExtractCode)
	if err != nil {
		m.finishTask(task, "failed", fmt.Sprintf("Failed to parse share: %v", err))
		return
	}

	// 2.2 获取目标目录已存在文件，用于去重跳过
	m.updateProgress(task, 35, "Checking", "正在检查目标目录是否存在同名文件...")
	targetID, err := driver.PrepareTargetPath(ctx, task.SavePath)
	existingMap := make(map[string]bool)
	if err == nil {
		existingFiles, _ := driver.ListFiles(ctx, targetID)
		for _, ef := range existingFiles {
			existingMap[ef.Name] = true
		}
	}

	// 2.3 日期、ID 过滤及同名去重
	renameProc := renamer.NewProcessor()
	var filteredIDs []string
	renameMap := make(map[string]string) // 记录 OriginalName -> NewName 的映射，用于后续重命名
	skippedCount := 0
	for _, f := range files {
		// 1. 检查日期过滤：如果设置了开始日期，且文件时间早于开始日期，则跳过
		if task.StartDate != nil && !f.UpdateTime.IsZero() && f.UpdateTime.Before(*task.StartDate) {
			continue
		}

		// 2. 检查 ID 截断
		isStartFile := task.StartFileID != "" && f.ID == task.StartFileID

		// 计算该文件经过重命名后的目标名称
		targetName := f.Name
		if task.Pattern != "" || task.Replacement != "" {
			opts := renamer.RenameOptions{
				TaskName:    task.Name,
				Pattern:     task.Pattern,
				Replacement: task.Replacement,
				FileName:    f.Name,
			}
			if newName, err := renameProc.Process(opts); err == nil {
				targetName = newName
			}
		}

		// 3. 检查同名跳过 (基于重命名后的名称进行检查)
		if existingMap[targetName] {
			skippedCount++
			if isStartFile {
				break
			}
			continue
		}

		filteredIDs = append(filteredIDs, f.ID)
		if targetName != f.Name {
			renameMap[f.Name] = targetName
		}

		if isStartFile {
			break
		}
	}

	if len(filteredIDs) == 0 {
		msg := "无新文件需要转存"
		if skippedCount > 0 {
			msg = fmt.Sprintf("无新文件需要转存 (已跳过 %d 个同名文件)", skippedCount)
		}
		m.finishTask(task, "success", msg)
		return
	}

	// 2.4 执行转存
	m.updateProgress(task, 65, "Saving", fmt.Sprintf("正在转存 %d 个文件 (已跳过 %d 个同名文件)...", len(filteredIDs), skippedCount))
	err = driver.SaveLink(ctx, task.ShareURL, task.ExtractCode, task.SavePath, filteredIDs)
	if err != nil {
		m.finishTask(task, "failed", err.Error())
		return
	}

	// 3. 执行重命名同步逻辑
	if len(renameMap) > 0 {
		m.updateProgress(task, 85, "Renaming", "正在同步执行网盘内重命名...")
		// 转存完成后，重新拉取目录以获取新产生的文件 ID
		targetFiles, err := driver.ListFiles(ctx, targetID)
		if err == nil {
			for _, tf := range targetFiles {
				if newName, ok := renameMap[tf.Name]; ok {
					log.Printf("[Worker:%d] 正在重命名: %s -> %s", task.ID, tf.Name, newName)
					_ = driver.RenameFile(ctx, tf.ID, newName)
				}
			}
		}
	}

	m.finishTask(task, "success", "Transfer completed successfully")
}

func (m *Manager) finishTask(task *db.Task, status, message string) {
	task.Status = status
	task.Message = message
	task.LastRun = time.Now()
	task.Percent = 100
	if status == "success" {
		task.Stage = "Success"
	} else {
		task.Stage = "Failed"
	}

	m.db.Model(task).Updates(map[string]interface{}{
		"status":   status,
		"message":  message,
		"last_run": task.LastRun,
		"percent":  task.Percent,
		"stage":    task.Stage,
	})
	log.Printf("Task %d finished with status: %s", task.ID, status)
	log.Printf("[PROGRESS:%d:100:%s:%s]", task.ID, task.Stage, message)
	utils.BroadcastTaskUpdate(task)
	utils.BroadcastStatsUpdate()
}
