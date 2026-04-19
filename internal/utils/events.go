package utils

import (
	"encoding/json"
	"log/slog"

	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// Event 定义了系统实时推送的事件结构
type Event struct {
	Type   string   `json:"type"`             // task_update, task_delete, stats_update
	Task   *db.Task `json:"task,omitempty"`   // 仅用于 task_update
	TaskID uint     `json:"taskId,omitempty"` // 仅用于 task_delete
}

// BroadcastTaskUpdate 推送任务状态更新（包含 ID、状态、进度、阶段等）
func BroadcastTaskUpdate(task *db.Task) {
	b, _ := json.Marshal(Event{Type: "task_update", Task: task})
	slog.Info("[EVENT:" + string(b) + "]")
}

// BroadcastTaskDelete 推送任务删除事件
func BroadcastTaskDelete(id uint) {
	b, _ := json.Marshal(Event{Type: "task_delete", TaskID: id})
	slog.Info("[EVENT:" + string(b) + "]")
}

// BroadcastStatsUpdate 通知前端刷新仪表盘统计数据
func BroadcastStatsUpdate() {
	b, _ := json.Marshal(Event{Type: "stats_update"})
	slog.Info("[EVENT:" + string(b) + "]")
}
