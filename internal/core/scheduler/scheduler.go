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
	cron     *cron.Cron
	EntryIDs map[uint]cron.EntryID
	mu       sync.RWMutex
}

var Global *Scheduler
var workerManager *worker.Manager

func Init(wm *worker.Manager) {
	workerManager = wm
	Global = New()
}

func New() *Scheduler {
	return &Scheduler{
		cron:     cron.New(cron.WithSeconds()), // 支持秒级
		EntryIDs: make(map[uint]cron.EntryID),
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

func (s *Scheduler) RemoveTask(taskID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entryID, exists := s.EntryIDs[taskID]; exists {
		s.cron.Remove(entryID)
		delete(s.EntryIDs, taskID)
		log.Printf("[Scheduler] Removed task %d", taskID)
	}
}

func (s *Scheduler) UpdateTask(taskID uint, cronExpr string) error {
	s.RemoveTask(taskID)

	if strings.TrimSpace(cronExpr) == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, err := s.cron.AddFunc(cronExpr, func() {
		var task db.Task
		if err := db.DB.Preload("Account").First(&task, taskID).Error; err != nil {
			log.Printf("[Scheduler] Task %d not found, removing from cron", taskID)
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

		log.Printf("[Scheduler] Triggering task %d", taskID)
		if workerManager != nil {
			workerManager.Submit(worker.Job{Task: &task})
		}
	})

	if err != nil {
		log.Printf("[Scheduler] Failed to add task %d: %v", taskID, err)
		return err
	}

	s.EntryIDs[taskID] = entryID
	log.Printf("[Scheduler] Added task %d with cron: %s", taskID, cronExpr)
	return nil
}
