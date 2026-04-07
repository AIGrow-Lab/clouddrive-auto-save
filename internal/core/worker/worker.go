package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// Job 代表一个待执行的转存任务
type Job struct {
	Task *db.Task
}

// Manager 负责管理 Worker 池和任务分发
type Manager struct {
	workers    int
	jobQueue   chan Job
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewManager(numWorkers int) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		workers:  numWorkers,
		jobQueue: make(chan Job, 100),
		ctx:      ctx,
		cancel:   cancel,
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
			m.execute(job)
		}
	}
}

func (m *Manager) execute(job Job) {
	task := job.Task
	log.Printf("Executing task: %s (ID: %d)", task.Name, task.ID)

	// 1. 更新任务状态为 running
	db.DB.Model(task).Update("status", "running")

	// 2. 获取对应的云盘驱动
	driver := core.GetDriver(&task.Account)
	if driver == nil {
		m.finishTask(task, "failed", fmt.Sprintf("Driver not found for platform: %s", task.Account.Platform))
		return
	}

	// 3. 执行转存逻辑
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Minute)
	defer cancel()

	err := driver.SaveLink(ctx, task.ShareURL, task.ExtractCode, task.SavePath)
	if err != nil {
		m.finishTask(task, "failed", err.Error())
		return
	}

	// 4. TODO: 执行重命名引擎逻辑

	m.finishTask(task, "success", "Transfer completed successfully")
}

func (m *Manager) finishTask(task *db.Task, status, message string) {
	task.Status = status
	task.Message = message
	task.LastRun = time.Now()
	
	db.DB.Model(task).Updates(map[string]interface{}{
		"status":   status,
		"message":  message,
		"last_run": task.LastRun,
	})
	log.Printf("Task %d finished with status: %s", task.ID, status)
}
