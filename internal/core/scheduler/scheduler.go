package scheduler

import (
	"log"
	"strings"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

type Scheduler struct {
	cron           *cron.Cron
	CustomEntryIDs map[uint]cron.EntryID
	GlobalEntryID  cron.EntryID
	mu             sync.RWMutex
}

var Global *Scheduler
var workerManager *worker.Manager

func Init(wm *worker.Manager) {
	workerManager = wm
	Global = New()
}

func New() *Scheduler {
	return &Scheduler{
		cron:           cron.New(cron.WithSeconds()), // 支持秒级
		CustomEntryIDs: make(map[uint]cron.EntryID),
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
	log.Println("[Scheduler] Started")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("[Scheduler] Stopped")
}

// ValidateCron 验证给定的 Cron 表达式是否有效
func ValidateCron(cronExpr string) error {
	if strings.TrimSpace(cronExpr) == "" {
		return nil
	}
	tempCron := cron.New(cron.WithSeconds())
	_, err := tempCron.AddFunc(cronExpr, func() {})
	return err
}

func (s *Scheduler) RemoveTask(taskID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entryID, exists := s.CustomEntryIDs[taskID]; exists {
		s.cron.Remove(entryID)
		delete(s.CustomEntryIDs, taskID)
		log.Printf("[Scheduler] Removed custom schedule for task %d", taskID)
	}
}

// UpdateGlobalSchedule 更新全局调度规则
func (s *Scheduler) UpdateGlobalSchedule(cronExpr string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.GlobalEntryID != 0 {
		s.cron.Remove(s.GlobalEntryID)
		s.GlobalEntryID = 0
		log.Println("[Scheduler] Removed old global schedule")
	}

	if !enabled || strings.TrimSpace(cronExpr) == "" {
		return nil
	}

	entryID, err := s.cron.AddFunc(cronExpr, func() {
		log.Println("[Scheduler] Triggering Global Schedule")
		var tasks []db.Task
		// 查出所有设置为 "跟随全局 (global)" 的任务
		if err := db.DB.Preload("Account").Where("schedule_mode = ?", "global").Find(&tasks).Error; err != nil {
			log.Printf("[Scheduler] Failed to fetch global tasks: %v", err)
			return
		}

		for _, task := range tasks {
			if task.Status == "running" {
				log.Printf("[Scheduler] Task %d is already running, skipping", task.ID)
				continue
			}
			if strings.Contains(task.Message, "[Fatal]") {
				log.Printf("[Scheduler] Task %d has fatal error, skipping", task.ID)
				continue
			}

			log.Printf("[Scheduler] Triggering task %d (Global Rule)", task.ID)
			if workerManager != nil {
				taskCopy := task // 避免闭包变量捕获问题
				workerManager.Submit(worker.Job{Task: &taskCopy})
			}
		}
	})

	if err != nil {
		log.Printf("[Scheduler] Failed to add global schedule: %v", err)
		return err
	}

	s.GlobalEntryID = entryID
	log.Printf("[Scheduler] Global schedule updated: %s (enabled: %v)", cronExpr, enabled)
	return nil
}

// UpdateTask 更新单个任务的自定义调度
func (s *Scheduler) UpdateTask(taskID uint, mode string, customCron string) error {
	s.RemoveTask(taskID)

	if mode != "custom" || strings.TrimSpace(customCron) == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, err := s.cron.AddFunc(customCron, func() {
		var task db.Task
		if err := db.DB.Preload("Account").First(&task, taskID).Error; err != nil {
			log.Printf("[Scheduler] Task %d not found, removing from custom cron", taskID)
			s.RemoveTask(taskID)
			return
		}

		if task.Status == "running" {
			log.Printf("[Scheduler] Task %d is already running, skipping", taskID)
			return
		}
		if strings.Contains(task.Message, "[Fatal]") {
			log.Printf("[Scheduler] Task %d has fatal error, skipping", taskID)
			return
		}

		log.Printf("[Scheduler] Triggering task %d (Custom Rule)", taskID)
		if workerManager != nil {
			workerManager.Submit(worker.Job{Task: &task})
		}
	})

	if err != nil {
		log.Printf("[Scheduler] Failed to add custom task %d: %v", taskID, err)
		return err
	}

	s.CustomEntryIDs[taskID] = entryID
	log.Printf("[Scheduler] Added custom task %d with cron: %s", taskID, customCron)
	return nil
}
