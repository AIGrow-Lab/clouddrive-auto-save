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
// forceEnabled 为 true 时忽略全局开关（用于手动扫描）
func (s *Scanner) ReloadConfig(forceEnabled bool) error {
	var enabled, apiURL, apiToken db.Setting

	db.DB.Where("key = ?", "openlist_enabled").First(&enabled)
	db.DB.Where("key = ?", "openlist_api_url").First(&apiURL)
	db.DB.Where("key = ?", "openlist_api_token").First(&apiToken)

	slog.Debug("OpenList 配置加载",
		"enabled", enabled.Value,
		"api_url", apiURL.Value,
		"has_token", apiToken.Value != "",
		"force_enabled", forceEnabled)

	s.mu.Lock()
	defer s.mu.Unlock()

	if !forceEnabled && enabled.Value != "true" {
		s.client = nil
		slog.Debug("OpenList 全局开关未开启，扫描已禁用")
		return nil
	}

	if apiURL.Value == "" || apiToken.Value == "" {
		s.client = nil
		slog.Debug("OpenList API 配置不完整，扫描已禁用")
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
		slog.Debug("OpenList client 未初始化，跳过扫描")
		return nil
	}

	slog.Info("触发 OpenList 扫描")
	if err := client.StartScan(ctx); err != nil {
		slog.Warn("OpenList 扫描失败", "error", err)
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
