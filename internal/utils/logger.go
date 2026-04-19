package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
)

// DashboardHandler 实现 slog.Handler 接口，将日志发送到仪表盘广播器
type DashboardHandler struct {
	slog.Handler
	broadcaster *Broadcaster
}

func (h *DashboardHandler) Handle(ctx context.Context, r slog.Record) error {
	// 0. 检查是否是纯数据事件 (以 [EVENT: 开头)
	isEvent := strings.HasPrefix(r.Message, "[EVENT:")

	// 1. 处理控制台输出 (Stdout)
	// 如果是普通日志，遵循 minLevel 过滤
	// 如果是事件日志，仅在 DEBUG 等级下才打印到控制台，避免生产环境噪音
	shouldPrintToConsole := false
	if isEvent {
		// 只有当前等级允许 DEBUG 时，才在控制台打印事件详情
		shouldPrintToConsole = h.Handler.Enabled(ctx, slog.LevelDebug)
	} else {
		// 普通日志正常检查等级
		shouldPrintToConsole = h.Handler.Enabled(ctx, r.Level)
	}

	if shouldPrintToConsole {
		if err := h.Handler.Handle(ctx, r); err != nil {
			return err
		}
	}

	// 2. 处理仪表盘广播 (Dashboard)
	// 事件日志 [EVENT:...] 必须始终广播，否则前端 UI 无法实时更新状态
	// 普通日志则遵循其本身的等级过滤
	shouldBroadcast := isEvent || h.Handler.Enabled(ctx, r.Level)

	if shouldBroadcast {
		msg := r.Message
		if !isEvent {
			// 普通日志添加等级前缀和结构化字段
			levelStr := r.Level.String()
			msg = fmt.Sprintf("[%s] %s", levelStr, r.Message)
			r.Attrs(func(a slog.Attr) bool {
				if a.Key != "" {
					msg = fmt.Sprintf("%s (%s=%v)", msg, a.Key, a.Value)
				}
				return true
			})
		}
		h.broadcaster.Broadcast(msg)
	}

	return nil
}

// InitLogger 初始化全局 slog
func InitLogger(minLevel slog.Level, out io.Writer) {
	if out == nil {
		out = os.Stdout
	}

	// 创建底层的 TextHandler
	baseHandler := slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: minLevel,
	})

	// 包装为 DashboardHandler
	handler := &DashboardHandler{
		Handler:     baseHandler,
		broadcaster: GlobalBroadcaster,
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	// 设置标准 log 库也输出到 slog (默认 INFO 级别)
	log.SetOutput(slog.NewLogLogger(handler, slog.LevelInfo).Writer())
	log.SetFlags(0) // 移除标准 log 的时间戳，因为 slog 会带
}
