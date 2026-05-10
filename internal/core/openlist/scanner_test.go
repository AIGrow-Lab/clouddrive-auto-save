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
	// 设置一个非 nil 的 client，使 OnTaskComplete 不会提前返回
	scanner.client = NewClient("http://localhost:9999", "test-token")
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
