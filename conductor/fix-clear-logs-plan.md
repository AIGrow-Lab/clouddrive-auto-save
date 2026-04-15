# 修复仪表盘日志清空不彻底问题

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 解决仪表盘点击“清空日志”后，刷新页面又会重新拉取旧日志的问题。通过在后端增加清空接口，实现前后端同步彻底清空日志。

**Architecture:** 
- 修改 `internal/utils/broadcaster.go`，增加 `ClearRecent` 方法清空历史数组。
- 修改 `internal/api/router.go`，新增 `DELETE /dashboard/logs/recent` 接口。
- 修改 `web/src/api/dashboard.js`，新增调用该清空接口的方法。
- 修改 `web/src/views/Dashboard.vue`，在点击清空按钮时同步调用后端接口。

**Tech Stack:** Go, Vue 3, Element Plus

---

### Task 1: 增加后端清空逻辑

**Files:**
- Modify: `internal/utils/broadcaster.go`
- Modify: `internal/api/router.go`

- [ ] **Step 1: 在 `broadcaster.go` 中增加 `ClearRecent` 方法**
```go
// ClearRecent 清空最近的历史日志
func (b *Broadcaster) ClearRecent() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.history = make([]string, 0, 50)
}
```

- [ ] **Step 2: 在 `router.go` 中增加接口和处理函数**
在 `api.GET("/dashboard/logs/recent", getRecentLogs)` 后添加：
```go
api.DELETE("/dashboard/logs/recent", clearRecentLogs)
```
并在文件末尾添加处理逻辑：
```go
func clearRecentLogs(c *gin.Context) {
	utils.GlobalBroadcaster.ClearRecent()
	c.PureJSON(http.StatusOK, gin.H{"message": "logs cleared"})
}
```

### Task 2: 完善前端清空交互

**Files:**
- Modify: `web/src/api/dashboard.js`
- Modify: `web/src/views/Dashboard.vue`

- [ ] **Step 1: 在 `dashboard.js` 中新增接口调用**
```javascript
export function clearLogsAPI() {
  return request({
    url: '/dashboard/logs/recent',
    method: 'delete'
  })
}
```

- [ ] **Step 2: 在 `Dashboard.vue` 中修改 `clearLogs` 逻辑**
引入 `clearLogsAPI`，并在 `clearLogs` 中调用。
```javascript
import { getStats, clearLogsAPI } from '../api/dashboard'

// ...

const clearLogs = async () => {
  try {
    await clearLogsAPI()
    logs.value = []
    ElMessage.success('日志已彻底清空')
  } catch (err) {
    console.error('清空日志失败:', err)
    ElMessage.error('清空日志失败')
  }
}
```

### Task 3: 验证与提交

- [ ] **Step 1: 编译测试**
  在 `web/` 目录下运行 `npm run build` 确保无语法错误。
- [ ] **Step 2: 自动提交并推送**
  按照功能逻辑分批 commit 并推送到远端 `main` 分支。
