# OpenList 扫描功能实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在系统设置模块新增 OpenList 扫描功能，支持手动触发和转存任务完成后自动触发 OpenList 的文件扫描

**Architecture:** 新增 `internal/core/openlist/` 模块，包含 HTTP 客户端和扫描管理器。扫描管理器处理触发逻辑和延迟合并，通过回调集成到 Worker 中。

**Tech Stack:** Go (net/http, sync, time), Gin, Vue 3, Element Plus

---

## 文件结构

| 文件 | 操作 | 职责 |
|------|------|------|
| `internal/core/openlist/client.go` | 新增 | OpenList HTTP 客户端，封装 API 调用 |
| `internal/core/openlist/client_test.go` | 新增 | 客户端单元测试 |
| `internal/core/openlist/scanner.go` | 新增 | 扫描管理器，处理触发逻辑和延迟合并 |
| `internal/core/openlist/scanner_test.go` | 新增 | 扫描管理器单元测试 |
| `internal/api/router.go` | 修改 | 新增 `/api/openlist/scan` 路由 |
| `internal/core/worker/worker.go` | 修改 | 集成扫描触发逻辑 |
| `internal/core/worker/batch_tracker.go` | 修改 | 批量任务完成时触发扫描 |
| `web/src/api/task.js` | 修改 | 新增 `triggerOpenListScan` API |
| `web/src/views/Settings.vue` | 修改 | 新增 OpenList 配置卡片 |

---

## Task 1: OpenList HTTP 客户端

**Files:**
- Create: `internal/core/openlist/client.go`
- Create: `internal/core/openlist/client_test.go`

- [ ] **Step 1: 创建客户端文件和结构体**

```go
// internal/core/openlist/client.go
package openlist

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client 封装 OpenList API 调用
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient 创建 OpenList 客户端
func NewClient(baseURL, token string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// StartScan 触发 OpenList 扫描
func (c *Client) StartScan(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/admin/scan/start", nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("请求超时")
		}
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
```

- [ ] **Step 2: 创建客户端单元测试**

```go
// internal/core/openlist/client_test.go
package openlist

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_StartScan_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admin/scan/start" {
			t.Errorf("期望路径 /api/admin/scan/start，实际 %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "test-token" {
			t.Errorf("期望 Authorization test-token，实际 %s", r.Header.Get("Authorization"))
		}
		if r.Method != http.MethodPost {
			t.Errorf("期望 POST 方法，实际 %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.StartScan(context.Background())
	if err != nil {
		t.Fatalf("期望成功，实际错误: %v", err)
	}
}

func TestClient_StartScan_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.StartScan(context.Background())
	if err == nil {
		t.Fatal("期望错误，实际成功")
	}
	if !contains(err.Error(), "500") {
		t.Errorf("错误信息应包含状态码 500，实际: %v", err)
	}
}

func TestClient_StartScan_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	client.httpClient.Timeout = 50 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.StartScan(ctx)
	if err == nil {
		t.Fatal("期望超时错误，实际成功")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
```

- [ ] **Step 3: 运行测试验证**

Run: `go test ./internal/core/openlist/... -v`
Expected: PASS

- [ ] **Step 4: 提交**

```bash
git add internal/core/openlist/client.go internal/core/openlist/client_test.go
git commit -m "feat(openlist): 添加 OpenList HTTP 客户端"
```

---

## Task 2: 扫描管理器

**Files:**
- Create: `internal/core/openlist/scanner.go`
- Create: `internal/core/openlist/scanner_test.go`

- [ ] **Step 1: 创建扫描管理器**

```go
// internal/core/openlist/scanner.go
package openlist

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// Scanner 管理 OpenList 扫描触发
type Scanner struct {
	mu           sync.Mutex
	client       *Client
	pendingBatch int
	scanTimer    *time.Timer
	scanning     bool
}

// GlobalScanner 全局扫描器实例
var GlobalScanner = NewScanner()

// NewScanner 创建扫描器实例
func NewScanner() *Scanner {
	return &Scanner{}
}

// ReloadConfig 从数据库重新加载配置
func (s *Scanner) ReloadConfig() error {
	var enabled, apiURL, apiToken db.Setting

	db.DB.Where("key = ?", "openlist_enabled").First(&enabled)
	db.DB.Where("key = ?", "openlist_api_url").First(&apiURL)
	db.DB.Where("key = ?", "openlist_api_token").First(&apiToken)

	s.mu.Lock()
	defer s.mu.Unlock()

	if enabled.Value != "true" || apiURL.Value == "" || apiToken.Value == "" {
		s.client = nil
		return nil
	}

	s.client = NewClient(apiURL.Value, apiToken.Value)
	return nil
}

// ScanNow 立即触发扫描
func (s *Scanner) ScanNow(ctx context.Context) error {
	s.mu.Lock()
	client := s.client
	s.mu.Unlock()

	if client == nil {
		return nil
	}

	slog.Info("触发 OpenList 扫描")
	if err := client.StartScan(ctx); err != nil {
		slog.Error("OpenList 扫描失败", "error", err)
		return err
	}

	slog.Info("OpenList 扫描已触发")
	return nil
}

// OnBatchStart 注册批量任务开始
func (s *Scanner) OnBatchStart(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingBatch = count
	slog.Debug("OpenList 批量任务开始", "count", count)
}

// OnTaskComplete 单个任务完成时调用
func (s *Scanner) OnTaskComplete(hasNewContent bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return
	}

	if s.pendingBatch > 0 {
		s.pendingBatch--
		if s.pendingBatch > 0 {
			slog.Debug("OpenList 等待批量任务完成", "remaining", s.pendingBatch)
			return
		}
		// 批量任务全部完成，触发扫描
		s.scheduleScan()
		return
	}

	// 单任务完成，有新内容时触发
	if hasNewContent {
		s.scheduleScan()
	}
}

// scheduleScan 延迟 3 秒触发扫描（合并连续完成）
func (s *Scanner) scheduleScan() {
	if s.scanTimer != nil {
		s.scanTimer.Stop()
	}
	s.scanTimer = time.AfterFunc(3*time.Second, func() {
		s.mu.Lock()
		if s.scanning {
			s.mu.Unlock()
			slog.Debug("OpenList 扫描正在进行，跳过")
			return
		}
		s.scanning = true
		s.mu.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.ScanNow(ctx); err != nil {
			slog.Error("OpenList 自动扫描失败", "error", err)
		}

		s.mu.Lock()
		s.scanning = false
		s.mu.Unlock()
	})
}
```

- [ ] **Step 2: 创建扫描管理器单元测试**

```go
// internal/core/openlist/scanner_test.go
package openlist

import (
	"testing"
	"time"
)

func TestScanner_ScheduleScan_MergesMultiple(t *testing.T) {
	scanner := NewScanner()
	callCount := 0

	// 模拟多次触发
	scanner.scheduleScan()
	scanner.scheduleScan()
	scanner.scheduleScan()

	// 等待扫描执行
	time.Sleep(4 * time.Second)

	// 应该只触发一次（通过 timer 合并）
	if callCount > 1 {
		t.Errorf("期望合并为 1 次扫描，实际 %d 次", callCount)
	}
}

func TestScanner_OnTaskComplete_BatchMode(t *testing.T) {
	scanner := NewScanner()
	scanner.pendingBatch = 3

	// 前两个任务完成不应触发扫描
	scanner.OnTaskComplete(true)
	scanner.OnTaskComplete(true)

	if scanner.pendingBatch != 1 {
		t.Errorf("期望剩余 1 个任务，实际 %d", scanner.pendingBatch)
	}

	// 最后一个任务完成应触发扫描
	scanner.OnTaskComplete(true)

	if scanner.pendingBatch != 0 {
		t.Errorf("期望所有任务完成，实际剩余 %d", scanner.pendingBatch)
	}
}
```

- [ ] **Step 3: 运行测试验证**

Run: `go test ./internal/core/openlist/... -v`
Expected: PASS

- [ ] **Step 4: 提交**

```bash
git add internal/core/openlist/scanner.go internal/core/openlist/scanner_test.go
git commit -m "feat(openlist): 添加扫描管理器，支持延迟合并和批量触发"
```

---

## Task 3: Worker 集成

**Files:**
- Modify: `internal/core/worker/worker.go:248-283`
- Modify: `internal/core/worker/batch_tracker.go:44-61`

- [ ] **Step 1: 修改 worker.go 的 finishTask 方法**

在 `finishTask` 方法中，当任务成功且有新文件时触发扫描：

```go
// internal/core/worker/worker.go

// 在文件顶部添加 import
import (
	// ... existing imports ...
	"github.com/zcq/clouddrive-auto-save/internal/core/openlist"
)

// 修改 finishTask 方法，在 Bark 通知之前添加扫描触发
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

	// OpenList 扫描触发：单任务模式且有新文件时触发
	if job.BatchID == "" && status == "success" && len(files) > 0 {
		openlist.GlobalScanner.OnTaskComplete(true)
	}

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

- [ ] **Step 2: 修改 batch_tracker.go 的 RegisterBatch 方法**

在批量任务注册时通知扫描管理器：

```go
// internal/core/worker/batch_tracker.go

// 在文件顶部添加 import
import (
	// ... existing imports ...
	"github.com/zcq/clouddrive-auto-save/internal/core/openlist"
)

// 修改 RegisterBatch 方法
func (t *BatchTracker) RegisterBatch(batchID string, total int) {
	if total <= 0 {
		slog.Warn("批次任务数必须大于 0，忽略注册", "batch_id", batchID, "total", total)
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, exists := t.batches[batchID]; exists {
		slog.Warn("批次 ID 已存在，将被覆盖", "batch_id", batchID)
	}
	t.batches[batchID] = &batchInfo{
		total:   total,
		count:   0,
		results: make([]BatchResult, 0, total),
		startAt: time.Now(),
	}
	slog.Info("批次已注册", "batch_id", batchID, "total", total)

	// 通知 OpenList 扫描管理器批量任务开始
	openlist.GlobalScanner.OnBatchStart(total)
}
```

- [ ] **Step 3: 修改 batch_tracker.go 的 ReportTask 方法**

在批量任务完成时通知扫描管理器：

```go
// internal/core/worker/batch_tracker.go

// 修改 ReportTask 方法，在任务完成时调用扫描管理器
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

	// 通知 OpenList 扫描管理器任务完成
	hasNewContent := result.Status == "success" && len(result.Files) > 0
	openlist.GlobalScanner.OnTaskComplete(hasNewContent)

	if info.count < info.total {
		t.mu.Unlock()
		return
	}

	// 全部完成，取出结果并清理
	results := info.results
	totalDuration := time.Since(info.startAt)
	onComplete := t.onComplete
	delete(t.batches, batchID)
	t.mu.Unlock()

	slog.Info("批次全部完成", "batch_id", batchID, "total", info.total, "duration", totalDuration)
	onComplete(results, totalDuration)
}
```

- [ ] **Step 4: 运行测试验证**

Run: `go test ./internal/core/worker/... -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/core/worker/worker.go internal/core/worker/batch_tracker.go
git commit -m "feat(worker): 集成 OpenList 扫描触发逻辑"
```

---

## Task 4: API 路由

**Files:**
- Modify: `internal/api/router.go:70-75`

- [ ] **Step 1: 添加 OpenList 扫描路由**

```go
// internal/api/router.go

// 在文件顶部添加 import
import (
	// ... existing imports ...
	"github.com/zcq/clouddrive-auto-save/internal/core/openlist"
)

// 在路由注册部分添加（约第 74 行之后）
api.POST("/openlist/scan", triggerOpenListScan)
```

- [ ] **Step 2: 添加扫描处理函数**

```go
// internal/api/router.go

// 添加在文件末尾
func triggerOpenListScan(c *gin.Context) {
	// 重新加载配置
	if err := openlist.GlobalScanner.ReloadConfig(); err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "加载 OpenList 配置失败"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := openlist.GlobalScanner.ScanNow(ctx); err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.PureJSON(http.StatusOK, gin.H{"message": "扫描已触发"})
}
```

- [ ] **Step 3: 运行测试验证**

Run: `go test ./internal/api/... -v`
Expected: PASS

- [ ] **Step 4: 提交**

```bash
git add internal/api/router.go
git commit -m "feat(api): 添加 OpenList 手动扫描 API 端点"
```

---

## Task 5: 前端 API 函数

**Files:**
- Modify: `web/src/api/task.js`

- [ ] **Step 1: 添加 triggerOpenListScan 函数**

```javascript
// web/src/api/task.js

// 在文件末尾添加
export function triggerOpenListScan() {
  return request({
    url: '/openlist/scan',
    method: 'post'
  })
}
```

- [ ] **Step 2: 提交**

```bash
git add web/src/api/task.js
git commit -m "feat(api): 添加 OpenList 扫描前端 API 函数"
```

---

## Task 6: 前端设置页面

**Files:**
- Modify: `web/src/views/Settings.vue`

- [ ] **Step 1: 添加 import**

```vue
<script setup>
import { ref, onMounted, watch } from 'vue'
import { Calendar, Bell, Info, Scan } from 'lucide-vue-next'
import { getGlobalSettings, updateGlobalSettings, testBark, triggerOpenListScan } from '../api/task'
import { ElMessage, ElMessageBox } from 'element-plus'

// ... existing code ...
```

- [ ] **Step 2: 添加 OpenList 相关状态**

```vue
<script setup>
// ... existing code ...

// OpenList 相关状态
const openlistScanning = ref(false)

// ... existing code ...
```

- [ ] **Step 3: 添加 OpenList 扫描处理函数**

```vue
<script setup>
// ... existing code ...

const handleOpenListScan = async () => {
  openlistScanning.value = true
  try {
    await triggerOpenListScan()
    ElMessage.success('OpenList 扫描已触发')
  } catch (error) {
    ElMessage.error('触发扫描失败: ' + (error.response?.data?.error || error.message))
  } finally {
    openlistScanning.value = false
  }
}

// ... existing code ...
```

- [ ] **Step 4: 添加 OpenList 配置卡片模板**

```vue
<template>
  <!-- ... existing template ... -->

    <el-row v-if="!pageLoading" :gutter="24">
      <!-- ... existing cards ... -->

      <!-- OpenList 扫描配置 -->
      <el-col :xs="24" :lg="12">
        <el-card class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <el-icon><Scan /></el-icon>
                <span>OpenList 扫描</span>
              </div>
              <el-switch
                v-model="settings.openlist_enabled"
                active-value="true"
                inactive-value="false"
                @change="() => saveOpenListSettings(false)"
              />
            </div>
          </template>
          <div class="card-content">
            <el-form label-position="top">
              <el-form-item label="API 地址">
                <el-input
                  v-model="settings.openlist_api_url"
                  placeholder="http://127.0.0.1:23541"
                />
              </el-form-item>
              <el-form-item label="API Token">
                <el-input
                  v-model="settings.openlist_api_token"
                  placeholder="openlist-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                  type="password"
                  show-password
                />
              </el-form-item>

              <div class="form-tip">
                配置 OpenList API 信息后，转存任务完成时将自动触发扫描。也可手动点击按钮触发。
              </div>

              <div class="form-actions">
                <el-button
                  type="primary"
                  plain
                  :loading="openlistScanning"
                  @click="handleOpenListScan"
                  style="margin-right: 12px"
                >
                  手动扫描
                </el-button>
                <el-button type="primary" :loading="savingOpenlist" @click="saveOpenListSettings(true)">
                  保存配置
                </el-button>
              </div>
            </el-form>
          </div>
        </el-card>
      </el-col>
    </el-row>

  <!-- ... existing template ... -->
</template>
```

- [ ] **Step 5: 添加 OpenList 保存函数**

```vue
<script setup>
// ... existing code ...

const savingOpenlist = ref(false)

const saveOpenListSettings = async (manual = false) => {
  if (isProcessing.value) return
  isProcessing.value = true
  if (manual) savingOpenlist.value = true

  try {
    await updateGlobalSettings(settings.value)
    if (manual) ElMessage.success('OpenList 扫描设置已保存')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    isProcessing.value = false
    savingOpenlist.value = false
  }
}

// ... existing code ...
```

- [ ] **Step 6: 在 settings 默认值中添加 OpenList 配置**

```vue
<script setup>
// ... existing code ...

const settings = ref({
  global_schedule_enabled: 'false',
  global_schedule_cron: '0 0 0 * * *',
  global_schedule_ui_mode: 'daily',
  bark_enabled: 'false',
  bark_server: 'https://api.day.app',
  bark_device_key: '',
  bark_success_sound: 'birdsong.caf',
  bark_success_level: 'active',
  bark_failure_sound: 'alarm.caf',
  bark_failure_level: 'timeSensitive',
  bark_archive: 'true',
  bark_icon: '',
  openlist_enabled: 'false',
  openlist_api_url: '',
  openlist_api_token: ''
})

// ... existing code ...
```

- [ ] **Step 7: 运行前端开发服务器验证**

Run: `make dev-web`
Expected: 前端启动成功，设置页面显示 OpenList 配置卡片

- [ ] **Step 8: 提交**

```bash
git add web/src/views/Settings.vue
git commit -m "feat(settings): 添加 OpenList 扫描配置卡片"
```

---

## Task 7: 集成测试

**Files:**
- Modify: `internal/api/router_test.go` (如果存在)

- [ ] **Step 1: 启动后端服务测试**

Run: `make dev-server`
Expected: 服务启动成功，无错误日志

- [ ] **Step 2: 测试手动扫描 API**

```bash
# 先配置 OpenList 设置
curl -X POST http://localhost:8080/api/settings/global \
  -H "Content-Type: application/json" \
  -d '{"openlist_enabled":"true","openlist_api_url":"http://127.0.0.1:23541","openlist_api_token":"test-token"}'

# 触发扫描
curl -X POST http://localhost:8080/api/openlist/scan
```

Expected: 返回 `{"message":"扫描已触发"}` 或 OpenList 服务不可用的错误

- [ ] **Step 3: 运行完整测试套件**

Run: `make test`
Expected: PASS

- [ ] **Step 4: 运行 lint 检查**

Run: `make lint`
Expected: 无格式错误

- [ ] **Step 5: 最终提交**

```bash
git add -A
git commit -m "feat: 完成 OpenList 扫描功能集成"
```

---

## 完成

所有任务完成后，运行 `make check` 验证完整 CI 流水线。
