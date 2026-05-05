package worker

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
	"github.com/zcq/clouddrive-auto-save/internal/core/renamer"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
	"gorm.io/gorm"
)

// Job 代表一个待执行的转存任务
type Job struct {
	Task    *db.Task
	BatchID string // 为空表示单任务执行
}

// Manager 负责管理 Worker 池和任务分发
type Manager struct {
	workers  int
	jobQueue chan Job
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	db       *gorm.DB
	tracker  *BatchTracker
}

func NewManager(numWorkers int, dbInst *gorm.DB) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		workers:  numWorkers,
		jobQueue: make(chan Job, 100),
		ctx:      ctx,
		cancel:   cancel,
		db:       dbInst,
		tracker:  NewBatchTracker(),
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

// RegisterBatch 注册一个批量执行批次
func (m *Manager) RegisterBatch(batchID string, total int) {
	m.tracker.RegisterBatch(batchID, total)
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
			m.execute(job)
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

func (m *Manager) execute(job Job) {
	task := job.Task
	startTime := time.Now()
	slog.Info("正在执行任务", "name", task.Name, "id", task.ID)
	m.updateProgress(task, 5, "Started", "任务已进入执行队列")

	// 1. 更新任务状态为 running
	m.db.Model(task).Update("status", "running")

	driver := core.GetDriver(&task.Account)
	if driver == nil {
		m.finishTask(job, "failed", "Driver not found", nil, startTime)
		return
	}

	// 2. 解析分享内容
	m.updateProgress(task, 15, "Parsing", "正在解析分享链接...")
	files, err := driver.ParseShare(m.ctx, task.ShareURL, task.ExtractCode)
	if err != nil {
		m.finishTask(job, "failed", "解析分享失败: "+err.Error(), nil, startTime)
		return
	}

	// 2.1 排序：按更新时间从新到旧
	sort.Slice(files, func(i, j int) bool {
		return files[i].UpdateTime.After(files[j].UpdateTime)
	})

	// 2.2 如果有起始文件，执行截断
	if task.StartFileID != "" {
		foundIdx := -1
		for i, f := range files {
			if f.ID == task.StartFileID {
				foundIdx = i
				break
			}
		}
		if foundIdx != -1 {
			// 仅保留该文件及其之后更新的文件 (即 0 到 foundIdx)
			files = files[:foundIdx+1]
			slog.Info("已应用起始文件截断", "task_id", task.ID, "start_file", task.StartFileName, "remaining_count", len(files))
		}
	}

	// 3. 列出目标目录文件，进行去重检查
	m.updateProgress(task, 35, "Checking", "正在检查目标目录是否存在同名文件...")
	targetID, err := driver.PrepareTargetPath(m.ctx, task.SavePath)
	if err != nil {
		m.finishTask(job, "failed", "准备目标路径失败: "+err.Error(), nil, startTime)
		return
	}

	existingFiles, err := driver.ListFiles(m.ctx, targetID)
	if err != nil {
		m.finishTask(job, "failed", "列出目标目录失败: "+err.Error(), nil, startTime)
		return
	}

	existingNames := make(map[string]bool)
	for _, f := range existingFiles {
		existingNames[f.Name] = true
	}

	// 4. 计算预览名并应用过滤（正则匹配 + 智能去重）
	processor := renamer.NewProcessor()
	var filteredIDs []string
	var savedFileNames []string
	var skipCount int
	var regexSkipCount int

	var compiledReg *regexp.Regexp
	if task.Pattern != "" {
		compiledReg, _ = regexp.Compile(task.Pattern)
	}

	for _, f := range files {
		// a. 正则匹配过滤
		if compiledReg != nil {
			if !compiledReg.MatchString(f.Name) {
				regexSkipCount++
				continue
			}
		}

		// b. 计算预期的新名字
		newName := f.Name
		if task.Replacement != "" {
			resName, err := processor.Process(renamer.RenameOptions{
				TaskName:    task.Name,
				FileName:    f.Name,
				Pattern:     task.Pattern,
				Replacement: task.Replacement,
			})
			if err == nil {
				newName = resName
			}
		}

		// c. 智能去重：拿新名比对
		if existingNames[newName] {
			skipCount++
			continue
		}

		filteredIDs = append(filteredIDs, f.ID)
		savedFileNames = append(savedFileNames, newName)
	}

	if len(filteredIDs) == 0 {
		msg := fmt.Sprintf("没有需要转存的文件 (跳过 %d 个同名文件", skipCount)
		if regexSkipCount > 0 {
			msg += fmt.Sprintf(", 过滤 %d 个不匹配文件", regexSkipCount)
		}
		msg += ")"
		m.finishTask(job, "success", msg, nil, startTime)
		return
	}

	m.updateProgress(task, 60, "Saving", fmt.Sprintf("正在转存 %d 个文件...", len(filteredIDs)))
	err = driver.SaveLink(m.ctx, task.ShareURL, task.ExtractCode, task.SavePath, filteredIDs)
	if err != nil {
		m.finishTask(job, "failed", "转存失败: "+err.Error(), nil, startTime)
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

	m.finishTask(job, "success", fmt.Sprintf("转存成功 (新增 %d 个文件, 跳过 %d 个同名文件)", len(filteredIDs), skipCount), savedFileNames, startTime)
}

func (m *Manager) finishTask(job Job, status, message string, files []string, startTime time.Time) {
	task := job.Task
	task.Status = status
	task.Message = message
	task.LastRun = time.Now()
	task.Percent = 100
	if status == "success" {
		task.Stage = "Success"
	} else {
		task.Stage = "Failed"
	}

	duration := time.Since(startTime)

	m.db.Model(task).Updates(map[string]interface{}{
		"status":   status,
		"message":  message,
		"last_run": task.LastRun,
		"percent":  task.Percent,
		"stage":    task.Stage,
	})
	slog.Info("任务完成", "id", task.ID, "status", status, "duration", duration)
	slog.Info(fmt.Sprintf("[PROGRESS:%d:100:%s:%s]", task.ID, task.Stage, message))
	utils.BroadcastTaskUpdate(task)
	utils.BroadcastStatsUpdate()

	// Bark 通知：区分单任务和批量模式
	if job.BatchID != "" {
		m.tracker.ReportTask(job.BatchID, BatchResult{
			TaskName: task.Name, Status: status,
			Message: message, Files: files, Duration: duration,
		})
	} else {
		notify.SendTaskNotification(task.Name, status, message, files, duration)
	}
}
