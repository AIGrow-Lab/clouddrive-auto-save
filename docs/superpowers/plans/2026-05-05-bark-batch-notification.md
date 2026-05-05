# Bark 批量通知集成实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现批量执行任务时的 Bark 汇总通知，单任务保持现有立即推送行为。

**Architecture:** Worker 层新增 BatchTracker 追踪批量执行状态，finishTask 通过 Job.BatchID 区分单任务/批量模式，批量完成时调用 notify.SendBatchNotification 发送汇总。

**Tech Stack:** Go, sync.Mutex, 现有 notify 包

---

## 文件结构

| 文件 | 职责 |
|------|------|
| `internal/core/worker/batch_tracker.go` | BatchTracker 实现：批次注册、结果收集、完成触发 |
| `internal/core/worker/batch_tracker_test.go` | BatchTracker 单元测试 |
| `internal/core/notify/batch.go` | SendBatchNotification：批量汇总消息构造与发送 |
| `internal/core/notify/batch_test.go` | 批量通知消息格式测试 |
| `internal/core/worker/worker.go` | 修改：Job 扩展 BatchID、Manager 新增 tracker、finishTask 逻辑变更 |
| `internal/api/router.go` | 修改：runAllTasks 生成 batchID 并注册批次 |

---

### Task 1: 实现 BatchTracker 核心逻辑

**Files:**
- Create: `internal/core/worker/batch_tracker.go`
- Test: `internal/core/worker/batch_tracker_test.go`

- [ ] **Step 1: 编写 BatchTracker 失败测试**

```go
// internal/core/worker/batch_tracker_test.go
package worker

import (
	"sync"
	"testing"
	"time"
)

func TestBatchTracker_RegisterAndReport(t *testing.T) {
	tracker := NewBatchTracker()
	tracker.RegisterBatch("batch_1", 2)

	var mu sync.Mutex
	var notified bool
	tracker.onComplete = func(results []BatchResult, totalDuration time.Duration) {
		mu.Lock()
		notified = true
		mu.Unlock()
	}

	tracker.ReportTask("batch_1", BatchResult{
		TaskName: "task1", Status: "success", Message: "ok", Files: []string{"a.mp4"}, Duration: 5 * time.Second,
	})

	mu.Lock()
	if notified {
		t.Fatal("should not notify after first task, batch has 2 tasks")
	}
	mu.Unlock()

	tracker.ReportTask("batch_1", BatchResult{
		TaskName: "task2", Status: "failed", Message: "err", Duration: 2 * time.Second,
	})

	mu.Lock()
	if !notified {
		t.Fatal("should notify after all tasks complete")
	}
	mu.Unlock()
}

func TestBatchTracker_SingleTaskBatch(t *testing.T) {
	tracker := NewBatchTracker()
	tracker.RegisterBatch("batch_2", 1)

	var results []BatchResult
	tracker.onComplete = func(r []BatchResult, _ time.Duration) {
		results = r
	}

	tracker.ReportTask("batch_2", BatchResult{
		TaskName: "only_task", Status: "success", Message: "done", Duration: 3 * time.Second,
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].TaskName != "only_task" {
		t.Errorf("expected task name 'only_task', got '%s'", results[0].TaskName)
	}
}

func TestBatchTracker_UnknownBatch(t *testing.T) {
	tracker := NewBatchTracker()
	// 不注册批次，直接上报——不应 panic
	tracker.ReportTask("nonexistent", BatchResult{
		TaskName: "ghost", Status: "success", Message: "", Duration: 0,
	})
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test -run TestBatchTracker ./internal/core/worker/ -v`
Expected: 编译失败，`NewBatchTracker`、`BatchResult`、`BatchTracker` 未定义

- [ ] **Step 3: 实现 BatchTracker**

```go
// internal/core/worker/batch_tracker.go
package worker

import (
	"log/slog"
	"sync"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
)

// BatchResult 收集单个任务的执行结果
type BatchResult struct {
	TaskName string
	Status   string // "success" | "failed"
	Message  string
	Files    []string
	Duration time.Duration
}

type batchInfo struct {
	total   int
	count   int
	results []BatchResult
	startAt time.Time
}

// BatchTracker 追踪批量执行状态，当批次内所有任务完成时触发汇总通知
type BatchTracker struct {
	mu         sync.Mutex
	batches    map[string]*batchInfo
	onComplete func(results []BatchResult, totalDuration time.Duration) // 可替换，便于测试
}

// NewBatchTracker 创建追踪器实例
func NewBatchTracker() *BatchTracker {
	t := &BatchTracker{
		batches: make(map[string]*batchInfo),
	}
	t.onComplete = t.defaultOnComplete
	return t
}

// RegisterBatch 注册一个新批次
func (t *BatchTracker) RegisterBatch(batchID string, total int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.batches[batchID] = &batchInfo{
		total:   total,
		count:   0,
		results: make([]BatchResult, 0, total),
		startAt: time.Now(),
	}
	slog.Info("批次已注册", "batch_id", batchID, "total", total)
}

// ReportTask 上报单个任务完成结果，当全部完成时触发汇总通知
func (t *BatchTracker) ReportTask(batchID string, result BatchResult) {
	t.mu.Lock()
	info, ok := t.batches[batchID]
	if !ok {
		t.mu.Unlock()
		slog.Warn("上报了未知批次", "batch_id", batchID)
		return
	}

	info.results = append(info.results, result)
	info.count++

	if info.count < info.total {
		t.mu.Unlock()
		return
	}

	// 全部完成，取出结果并清理
	results := info.results
	totalDuration := time.Since(info.startAt)
	delete(t.batches, batchID)
	t.mu.Unlock()

	slog.Info("批次全部完成", "batch_id", batchID, "total", info.total, "duration", totalDuration)
	t.onComplete(results, totalDuration)
}

// defaultOnComplete 默认完成回调：转换类型并调用 notify 发送汇总
func (t *BatchTracker) defaultOnComplete(results []BatchResult, totalDuration time.Duration) {
	notifyResults := make([]notify.BatchResult, len(results))
	for i, r := range results {
		notifyResults[i] = notify.BatchResult{
			TaskName: r.TaskName,
			Status:   r.Status,
			Message:  r.Message,
			Files:    r.Files,
			Duration: r.Duration,
		}
	}
	notify.SendBatchNotification(notifyResults, totalDuration)
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test -run TestBatchTracker ./internal/core/worker/ -v`
Expected: PASS — 三个测试全部通过

- [ ] **Step 5: 提交**

```bash
git add internal/core/worker/batch_tracker.go internal/core/worker/batch_tracker_test.go
git commit -m "feat(worker): 新增 BatchTracker 批量执行状态追踪器"
```

---

### Task 2: 实现 SendBatchNotification

**Files:**
- Create: `internal/core/notify/batch.go`
- Test: `internal/core/notify/batch_test.go`

- [ ] **Step 1: 编写消息格式测试**

```go
// internal/core/notify/batch_test.go
package notify

import (
	"strings"
	"testing"
	"time"
)

func TestBuildBatchBody_AllSuccess(t *testing.T) {
	results := []BatchResult{
		{TaskName: "任务A", Status: "success", Message: "转存成功 (新增 5 个文件)", Files: []string{"a.mp4", "b.mp4"}, Duration: 45 * time.Second},
		{TaskName: "任务B", Status: "success", Message: "转存成功 (新增 3 个文件)", Files: []string{"c.mkv"}, Duration: 32 * time.Second},
	}
	body := buildBatchBody(results, 2*time.Minute)

	if !strings.Contains(body, "总耗时") {
		t.Error("body should contain '总耗时'")
	}
	if !strings.Contains(body, "✅ 任务A") {
		t.Error("body should contain success icon for 任务A")
	}
	if !strings.Contains(body, "转存文件列表") {
		t.Error("body should contain file list section")
	}
	if !strings.Contains(body, "a.mp4") {
		t.Error("body should list file names")
	}
}

func TestBuildBatchBody_MixedResults(t *testing.T) {
	results := []BatchResult{
		{TaskName: "任务A", Status: "success", Message: "ok", Files: []string{"a.mp4"}, Duration: 10 * time.Second},
		{TaskName: "任务B", Status: "failed", Message: "解析分享失败: 链接过期", Duration: 2 * time.Second},
	}
	body := buildBatchBody(results, 30*time.Second)

	if !strings.Contains(body, "❌ 任务B") {
		t.Error("body should contain failure icon for 任务B")
	}
	if !strings.Contains(body, "链接过期") {
		t.Error("body should contain failure reason")
	}
}

func TestBuildBatchTitle_AllSuccess(t *testing.T) {
	results := []BatchResult{
		{TaskName: "A", Status: "success"},
		{TaskName: "B", Status: "success"},
	}
	title := buildBatchTitle(results)
	if title != "📊 批量转存完成: 全部 2 个任务成功" {
		t.Errorf("unexpected title: %s", title)
	}
}

func TestBuildBatchTitle_PartialFailure(t *testing.T) {
	results := []BatchResult{
		{TaskName: "A", Status: "success"},
		{TaskName: "B", Status: "failed"},
		{TaskName: "C", Status: "success"},
	}
	title := buildBatchTitle(results)
	if title != "📊 批量转存完成: 2成功 / 1失败" {
		t.Errorf("unexpected title: %s", title)
	}
}

func TestBuildBatchBody_FileListTruncation(t *testing.T) {
	files := make([]string, 25)
	for i := range files {
		files[i] = "file.mp4"
	}
	results := []BatchResult{
		{TaskName: "BigTask", Status: "success", Message: "ok", Files: files, Duration: 10 * time.Second},
	}
	body := buildBatchBody(results, 10*time.Second)

	if !strings.Contains(body, "等共 25 个文件") {
		t.Error("body should truncate file list and show total count")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test -run TestBuildBatch ./internal/core/notify/ -v`
Expected: 编译失败，`BatchResult`、`buildBatchBody`、`buildBatchTitle` 未定义

- [ ] **Step 3: 实现 SendBatchNotification**

```go
// internal/core/notify/batch.go
package notify

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// BatchResult 批量任务结果
type BatchResult struct {
	TaskName string
	Status   string
	Message  string
	Files    []string
	Duration time.Duration
}

// buildBatchTitle 构造批量通知标题
func buildBatchTitle(results []BatchResult) string {
	total := len(results)
	successCount := 0
	for _, r := range results {
		if r.Status == "success" {
			successCount++
		}
	}
	failCount := total - successCount

	if failCount == 0 {
		return fmt.Sprintf("📊 批量转存完成: 全部 %d 个任务成功", total)
	}
	return fmt.Sprintf("📊 批量转存完成: %d成功 / %d失败", successCount, failCount)
}

// buildBatchBody 构造批量通知正文
func buildBatchBody(results []BatchResult, totalDuration time.Duration) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("总耗时: %s\n", formatDuration(totalDuration)))

	for _, r := range results {
		icon := "✅"
		if r.Status == "failed" {
			icon = "❌"
		}
		sb.WriteString(fmt.Sprintf("\n%s %s - %s - 耗时 %s", icon, r.TaskName, r.Message, formatDuration(r.Duration)))
	}

	// 收集所有文件
	var allFiles []string
	for _, r := range results {
		allFiles = append(allFiles, r.Files...)
	}

	if len(allFiles) > 0 {
		sb.WriteString("\n\n转存文件列表:")
		maxFiles := 20
		for i, f := range allFiles {
			if i >= maxFiles {
				sb.WriteString(fmt.Sprintf("\n... 等共 %d 个文件", len(allFiles)))
				break
			}
			sb.WriteString(fmt.Sprintf("\n- %s", f))
		}
	}

	return sb.String()
}

// formatDuration 格式化耗时，秒级精度
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}

// SendBatchNotification 发送批量执行汇总通知
func SendBatchNotification(results []BatchResult, totalDuration time.Duration) {
	var enabledSetting, serverSetting, keySetting, iconSetting, archiveSetting db.Setting

	db.DB.Where("key = ?", "bark_enabled").First(&enabledSetting)
	if enabledSetting.Value != "true" {
		return
	}

	db.DB.Where("key = ?", "bark_server").First(&serverSetting)
	db.DB.Where("key = ?", "bark_device_key").First(&keySetting)
	db.DB.Where("key = ?", "bark_icon").First(&iconSetting)
	db.DB.Where("key = ?", "bark_archive").First(&archiveSetting)

	server := serverSetting.Value
	key := keySetting.Value
	if key == "" {
		return
	}

	// 判断是否有失败任务
	hasFailure := false
	for _, r := range results {
		if r.Status == "failed" {
			hasFailure = true
			break
		}
	}

	// 根据成功/失败选择级别和铃声
	var level, sound string
	var levelSetting, soundSetting db.Setting
	if hasFailure {
		level = "timeSensitive"
		sound = "alarm.caf"
		db.DB.Where("key = ?", "bark_failure_level").First(&levelSetting)
		if levelSetting.Value != "" {
			level = levelSetting.Value
		}
		db.DB.Where("key = ?", "bark_failure_sound").First(&soundSetting)
		if soundSetting.Value != "" && soundSetting.Value != "default" {
			sound = soundSetting.Value
		}
	} else {
		level = "active"
		sound = "birdsong.caf"
		db.DB.Where("key = ?", "bark_success_level").First(&levelSetting)
		if levelSetting.Value != "" {
			level = levelSetting.Value
		}
		db.DB.Where("key = ?", "bark_success_sound").First(&soundSetting)
		if soundSetting.Value != "" && soundSetting.Value != "default" {
			sound = soundSetting.Value
		}
	}

	title := buildBatchTitle(results)
	body := buildBatchBody(results, totalDuration)
	icon := iconSetting.Value
	archive := archiveSetting.Value

	go func() {
		if err := SendBarkDirect(server, key, title, body, level, sound, icon, archive); err != nil {
			slog.Error("发送 Bark 批量通知失败", "err", err)
		}
	}()
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test -run TestBuildBatch ./internal/core/notify/ -v`
Expected: PASS — 四个测试全部通过

- [ ] **Step 5: 提交**

```bash
git add internal/core/notify/batch.go internal/core/notify/batch_test.go
git commit -m "feat(notify): 新增 SendBatchNotification 批量汇总通知"
```

---

### Task 3: 改造 Worker 层支持批量模式

**Files:**
- Modify: `internal/core/worker/worker.go:21-23,26-33,35-44,60-63,92-95,239-266`

- [ ] **Step 1: 扩展 Job 结构体**

在 `internal/core/worker/worker.go:21-23`，修改 Job 结构体：

```go
// Job 代表一个待执行的转存任务
type Job struct {
	Task    *db.Task
	BatchID string // 为空表示单任务执行
}
```

- [ ] **Step 2: Manager 新增 tracker 字段**

在 `internal/core/worker/worker.go:26-33`，修改 Manager 结构体：

```go
// Manager 负责管理 Worker 池 and 任务分发
type Manager struct {
	workers  int
	jobQueue chan Job
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	db       *gorm.DB
	tracker  *BatchTracker
}
```

- [ ] **Step 3: NewManager 初始化 tracker**

在 `internal/core/worker/worker.go:35-44`，修改 NewManager：

```go
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
```

- [ ] **Step 4: 修改 Submit 方法签名**

在 `internal/core/worker/worker.go:60-63`，修改 Submit：

```go
// Submit 提交一个任务
func (m *Manager) Submit(job Job) {
	m.jobQueue <- job
}
```

签名不变，但调用方现在传入的 Job 包含 BatchID。

- [ ] **Step 5: 修改 worker 方法传递 Job**

在 `internal/core/worker/worker.go:65-77`，修改 worker 方法：

```go
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
```

- [ ] **Step 6: 修改 execute 方法接收 Job**

在 `internal/core/worker/worker.go:92`，修改 execute 签名和开头：

```go
func (m *Manager) execute(job Job) {
	task := job.Task
	startTime := time.Now()
	slog.Info("正在执行任务", "name", task.Name, "id", task.ID)
	m.updateProgress(task, 5, "Started", "任务已进入执行队列")

	// 1. 更新任务状态为 running
	m.db.Model(task).Update("status", "running")
```

execute 内部所有对 `m.finishTask(task, ...)` 的调用改为 `m.finishTask(job, ...)`。具体位置（约 6 处调用）：

```go
// 原：m.finishTask(task, "failed", "Driver not found", nil, startTime)
// 改：
m.finishTask(job, "failed", "Driver not found", nil, startTime)

// 原：m.finishTask(task, "failed", "解析分享失败: "+err.Error(), nil, startTime)
// 改：
m.finishTask(job, "failed", "解析分享失败: "+err.Error(), nil, startTime)

// 原：m.finishTask(task, "failed", "准备目标路径失败: "+err.Error(), nil, startTime)
// 改：
m.finishTask(job, "failed", "准备目标路径失败: "+err.Error(), nil, startTime)

// 原：m.finishTask(task, "failed", "列出目标目录失败: "+err.Error(), nil, startTime)
// 改：
m.finishTask(job, "failed", "列出目标目录失败: "+err.Error(), nil, startTime)

// 原：m.finishTask(task, "success", msg, nil, startTime)
// 改：
m.finishTask(job, "success", msg, nil, startTime)

// 原：m.finishTask(task, "failed", "转存失败: "+err.Error(), nil, startTime)
// 改：
m.finishTask(job, "failed", "转存失败: "+err.Error(), nil, startTime)

// 原：m.finishTask(task, "success", fmt.Sprintf("转存成功 ..."), savedFileNames, startTime)
// 改：
m.finishTask(job, "success", fmt.Sprintf("转存成功 (新增 %d 个文件, 跳过 %d 个同名文件)", len(filteredIDs), skipCount), savedFileNames, startTime)
```

- [ ] **Step 7: 改造 finishTask 方法**

在 `internal/core/worker/worker.go:239-266`，替换整个 finishTask 方法：

```go
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
```

- [ ] **Step 8: 运行现有测试确保不破坏**

Run: `go test ./internal/core/worker/ -v`
Expected: 所有现有测试通过（现有测试调用 m.execute(&task)，需要适配为 m.execute(Job{Task: &task})）

注意：现有测试中 `m.execute(&task)` 需要改为 `m.execute(Job{Task: &task})`。修改 `internal/core/worker/worker_test.go` 中所有 `m.execute(&task)` 为 `m.execute(Job{Task: &task})`。

- [ ] **Step 9: 提交**

```bash
git add internal/core/worker/worker.go internal/core/worker/worker_test.go
git commit -m "refactor(worker): 改造 finishTask 支持批量模式，Job 携带 BatchID"
```

---

### Task 4: 改造 API 层 runAllTasks

**Files:**
- Modify: `internal/api/router.go:400-436`

- [ ] **Step 1: 修改 runAllTasks 生成 batchID**

在 `internal/api/router.go` 的 `runAllTasks` 函数中，修改提交循环部分：

```go
func runAllTasks(c *gin.Context) {
	slog.Info("请求批量运行所有任务")

	var tasks []db.Task
	err := db.DB.Preload("Account").
		Where("status != ?", "running").
		Where("message NOT LIKE ? OR message IS NULL", "%[Fatal]%").
		Find(&tasks).Error

	if err != nil {
		slog.Error("获取批量运行任务列表失败", "error", err)
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tasks"})
		return
	}

	if len(tasks) == 0 {
		c.PureJSON(http.StatusOK, gin.H{"message": "没有可执行的任务", "count": 0})
		return
	}

	// 生成批次 ID 并注册
	batchID := fmt.Sprintf("batch_%d", time.Now().UnixMilli())
	WorkerManager.RegisterBatch(batchID, len(tasks))

	count := 0
	for i := range tasks {
		task := &tasks[i]
		task.Status = "running"
		task.Stage = "Started"
		db.DB.Model(task).Updates(map[string]interface{}{
			"status": "running",
			"stage":  "Started",
		})
		utils.BroadcastTaskUpdate(task)

		WorkerManager.Submit(worker.Job{Task: task, BatchID: batchID})
		count++
	}

	utils.BroadcastStatsUpdate()
	slog.Info("批量运行任务提交完成", "batch_id", batchID, "total_triggered", count)
	c.PureJSON(http.StatusOK, gin.H{"message": fmt.Sprintf("批量执行已开启，共触发 %d 个任务", count), "count": count})
}
```

- [ ] **Step 2: 运行 go vet 检查**

Run: `go vet ./internal/api/...`
Expected: 无错误

- [ ] **Step 3: 提交**

```bash
git add internal/api/router.go
git commit -m "feat(api): runAllTasks 生成 batchID 支持批量通知"
```

---

### Task 5: 集成验证与全量测试

**Files:**
- Test: 运行全量测试

- [ ] **Step 1: 运行全部单元测试**

Run: `go test -race ./internal/... -v`
Expected: 所有测试通过，无 race condition

- [ ] **Step 2: 运行 go vet**

Run: `go vet ./...`
Expected: 无错误

- [ ] **Step 3: 运行 lint**

Run: `make lint`
Expected: 无格式问题

- [ ] **Step 4: 手动启动验证编译**

Run: `go build ./cmd/server/`
Expected: 编译成功

- [ ] **Step 5: 提交（如有修复）**

如有修复则提交，否则跳过。
