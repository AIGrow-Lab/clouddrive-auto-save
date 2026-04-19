package worker

import (
	"context"
	"fmt"
	"log/slog"
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
	for i := 1; i <= m.workers; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}
}

// Stop 停止所有 Worker
func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
}

// Submit 提交一个任务
func (m *Manager) Submit(job Job) {
	m.jobQueue <- job
}

func (m *Manager) worker(id int) {
	defer m.wg.Done()
	slog.Info("Worker 启动", "id", id)
	for {
		select {
		case <-m.ctx.Done():
			slog.Info("Worker 正在停止", "id", id)
			return
		case job := <-m.jobQueue:
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
	slog.Info(fmt.Sprintf("[PROGRESS:%d:%d:%s:%s]", task.ID, percent, stage, message))
	utils.BroadcastTaskUpdate(task)
}

func (m *Manager) execute(task *db.Task) {
	slog.Info("正在执行任务", "name", task.Name, "id", task.ID)
	m.updateProgress(task, 5, "Started", "任务已进入执行队列")

	// 1. 更新任务状态为 running
	m.db.Model(task).Update("status", "running")

	driver := core.GetDriver(&task.Account)
	if driver == nil {
		m.finishTask(task, "failed", "Driver not found")
		return
	}

	// 2. 解析分享内容
	m.updateProgress(task, 15, "Parsing", "正在解析分享链接...")
	files, err := driver.ParseShare(m.ctx, task.ShareURL, task.ExtractCode)
	if err != nil {
		m.finishTask(task, "failed", "解析分享失败: "+err.Error())
		return
	}

	// 3. 列出目标目录文件，进行去重检查
	m.updateProgress(task, 35, "Checking", "正在检查目标目录是否存在同名文件...")
	targetID, err := driver.PrepareTargetPath(m.ctx, task.SavePath)
	if err != nil {
		m.finishTask(task, "failed", "准备目标路径失败: "+err.Error())
		return
	}

	existingFiles, err := driver.ListFiles(m.ctx, targetID)
	if err != nil {
		m.finishTask(task, "failed", "列出目标目录失败: "+err.Error())
		return
	}

	existingNames := make(map[string]bool)
	for _, f := range existingFiles {
		existingNames[f.Name] = true
	}

	// 4. 执行转存
	var filteredIDs []string
	var skipCount int
	for _, f := range files {
		if existingNames[f.Name] {
			skipCount++
			continue
		}
		filteredIDs = append(filteredIDs, f.ID)
	}

	if len(filteredIDs) == 0 {
		m.finishTask(task, "success", fmt.Sprintf("无新文件需要转存 (已跳过 %d 个同名文件)", skipCount))
		return
	}

	m.updateProgress(task, 60, "Saving", fmt.Sprintf("正在转存 %d 个文件...", len(filteredIDs)))
	err = driver.SaveLink(m.ctx, task.ShareURL, task.ExtractCode, task.SavePath, filteredIDs)
	if err != nil {
		m.finishTask(task, "failed", "转存失败: "+err.Error())
		return
	}

	// 5. 检查是否需要重命名 (如果有规则)
	if task.Pattern != "" && task.Replacement != "" {
		m.updateProgress(task, 85, "Renaming", "转存成功，正在执行重命名...")
		// 再次列出文件，找到刚才存入的文件进行重命名
		newFiles, _ := driver.ListFiles(m.ctx, targetID)
		processor := renamer.NewProcessor()
		for _, tf := range newFiles {
			newName, err := processor.Process(renamer.RenameOptions{
				TaskName:    task.Name,
				FileName:    tf.Name,
				Pattern:     task.Pattern,
				Replacement: task.Replacement,
			})
			if err == nil && newName != tf.Name {
				slog.Info("正在执行重命名", "task_id", task.ID, "old_name", tf.Name, "new_name", newName)
				_ = driver.RenameFile(m.ctx, tf.ID, newName)
			}
		}
	}

	m.finishTask(task, "success", fmt.Sprintf("转存成功 (新增 %d 个文件, 跳过 %d 个同名文件)", len(filteredIDs), skipCount))
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
	slog.Info("任务完成", "id", task.ID, "status", status)
	slog.Info(fmt.Sprintf("[PROGRESS:%d:100:%s:%s]", task.ID, task.Stage, message))
	utils.BroadcastTaskUpdate(task)
	utils.BroadcastStatsUpdate()
}
