package worker

import (
	"log/slog"
	"sync"
	"time"
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
	onComplete := t.onComplete // 在锁内拷贝引用，避免数据竞争
	delete(t.batches, batchID)
	t.mu.Unlock()

	slog.Info("批次全部完成", "batch_id", batchID, "total", info.total, "duration", totalDuration)
	onComplete(results, totalDuration)
}

// defaultOnComplete 默认完成回调
// TODO: Task 2 中将对接 notify.SendBatchNotification 实现真正的汇总推送
func (t *BatchTracker) defaultOnComplete(results []BatchResult, totalDuration time.Duration) {
	slog.Info("批次汇总（回调未对接）", "task_count", len(results), "duration", totalDuration)
}
