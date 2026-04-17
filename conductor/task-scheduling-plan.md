# Task Scheduling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a hot-pluggable cron-based task scheduler to automatically run cloud drive save tasks.

**Architecture:** A new `scheduler` package wrapping `robfig/cron/v3` will manage in-memory cron jobs synced with the SQLite database. The API layer will update the scheduler on task changes. The frontend will provide a UI to set predefined or custom cron expressions.

**Tech Stack:** Go, Gin, GORM, `robfig/cron/v3`, Vue 3, Element Plus.

---

### Task 1: Add Cron Dependency and Create Scheduler Core

**Files:**
- Modify: `go.mod`, `go.sum`
- Create: `internal/core/scheduler/scheduler.go`
- Test: `internal/core/scheduler/scheduler_test.go`

- [ ] **Step 1: Add dependency**
Run: `go get github.com/robfig/cron/v3`

- [ ] **Step 2: Write the failing test**
Create `internal/core/scheduler/scheduler_test.go`:
```go
package scheduler

import (
	"testing"
)

func TestScheduler_AddAndRemoveTask(t *testing.T) {
	s := New()
	s.Start()
	defer s.Stop()

	err := s.UpdateTask(1, "* * * * *")
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	if len(s.EntryIDs) != 1 {
		t.Errorf("Expected 1 task in scheduler, got %d", len(s.EntryIDs))
	}

	s.RemoveTask(1)
	if len(s.EntryIDs) != 0 {
		t.Errorf("Expected 0 tasks after removal, got %d", len(s.EntryIDs))
	}
}
```

- [ ] **Step 3: Run test to verify it fails**
Run: `go test ./internal/core/scheduler -v`
Expected: FAIL (package does not exist or missing methods)

- [ ] **Step 4: Write minimal implementation**
Create `internal/core/scheduler/scheduler.go`:
```go
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
```

- [ ] **Step 5: Run test to verify it passes**
Run: `go test ./internal/core/scheduler -v`
Expected: PASS 

- [ ] **Step 6: Commit**
Run: `git add go.mod go.sum internal/core/scheduler`
Run: `git commit -m "feat(scheduler): add cron based task scheduler core"`

### Task 2: Integrate Scheduler with API and App Lifecycle

**Files:**
- Modify: `cmd/server/main.go`
- Modify: `internal/api/router.go`

- [ ] **Step 1: Write integration updates**

Modify `cmd/server/main.go` to initialize and load existing tasks:
```go
// ... existing imports
// add: "github.com/zcq/clouddrive-auto-save/internal/core/scheduler"

// Inside main() after api.WorkerManager.Start()
scheduler.Init(api.WorkerManager)
scheduler.Global.Start()
defer scheduler.Global.Stop()

// Load existing tasks
var tasks []db.Task
db.DB.Find(&tasks)
for _, t := range tasks {
	if t.Cron != "" {
		scheduler.Global.UpdateTask(t.ID, t.Cron)
	}
}
```

Modify `internal/api/router.go` to update scheduler on task changes:
Add import `"github.com/zcq/clouddrive-auto-save/internal/core/scheduler"`.

In `createTask` func, after `db.DB.Create(&task)`:
```go
	if task.Cron != "" {
		scheduler.Global.UpdateTask(task.ID, task.Cron)
	}
```

In `updateTask` func, after `db.DB.Model(&task).Updates(updateData)`:
```go
	// 刷新调度器
	scheduler.Global.UpdateTask(task.ID, task.Cron)
```

In `deleteTask` func, before `db.DB.Delete(&db.Task{}, id)`:
```go
	idNum, _ := strconv.Atoi(id)
	scheduler.Global.RemoveTask(uint(idNum))
```

- [ ] **Step 2: Build and verify**
Run: `go build -o bin/ucas cmd/server/main.go`
Expected: Successful build.

- [ ] **Step 3: Commit**
Run: `git add cmd/server/main.go internal/api/router.go`
Run: `git commit -m "feat(api): integrate scheduler with task lifecycle"`

### Task 3: Frontend UI for Task Scheduling

**Files:**
- Modify: `web/src/views/Tasks.vue`

- [ ] **Step 1: Add Cron UI components**
In `Tasks.vue` template, inside the `<el-dialog>` for editing tasks, before the `<el-row>` with `start_file_id`:
```html
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="定时执行 (Cron)">
              <div style="display: flex; align-items: center; gap: 15px; width: 100%;">
                <el-switch v-model="enableCron" active-text="开启" inactive-text="关闭" @change="handleCronSwitch" />
                <el-select v-if="enableCron" v-model="cronPreset" placeholder="选择预设频率" style="width: 180px" @change="handleCronPreset">
                  <el-option label="每小时" value="0 0 * * * *" />
                  <el-option label="每 6 小时" value="0 0 */6 * * *" />
                  <el-option label="每天凌晨 2 点" value="0 0 2 * * *" />
                  <el-option label="每周一凌晨 2 点" value="0 0 2 * * 1" />
                  <el-option label="自定义" value="custom" />
                </el-select>
                <el-input v-if="enableCron && cronPreset === 'custom'" v-model="form.cron" placeholder="Cron 表达式 (秒 分 时 日 月 周)" style="flex: 1" />
              </div>
            </el-form-item>
          </el-col>
        </el-row>
```

- [ ] **Step 2: Update Script Setup**
In `<script setup>` of `Tasks.vue`:
```javascript
const enableCron = ref(false)
const cronPreset = ref('')

const handleCronSwitch = (val) => {
  if (!val) {
    form.value.cron = ''
    cronPreset.value = ''
  } else {
    cronPreset.value = '0 0 * * * *'
    form.value.cron = '0 0 * * * *'
  }
}

const handleCronPreset = (val) => {
  if (val !== 'custom') {
    form.value.cron = val
  }
}
```

In `openAddDialog`:
```javascript
  enableCron.value = false
  cronPreset.value = ''
```

In `handleEdit`:
```javascript
  const isCustom = !['0 0 * * * *', '0 0 */6 * * *', '0 0 2 * * *', '0 0 2 * * 1'].includes(row.cron)
  enableCron.value = !!row.cron
  cronPreset.value = row.cron ? (isCustom ? 'custom' : row.cron) : ''
```

- [ ] **Step 3: Display Cron in Table**
Add a column to display the cron settings in the table:
```html
        <el-table-column prop="cron" label="定时规则" width="120" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag size="small" type="info" v-if="row.cron"><el-icon><Clock /></el-icon> {{ row.cron }}</el-tag>
            <span v-else class="empty-text">手动</span>
          </template>
        </el-table-column>
```
Remember to add `Clock` to the `lucide-vue-next` imports.

- [ ] **Step 4: Build and Verify**
Check if the UI appears correctly, updates state appropriately, and displays in the table.

- [ ] **Step 5: Commit**
Run: `git add web/src/views/Tasks.vue`
Run: `git commit -m "feat(ui): add visual cron scheduler configuration"`
