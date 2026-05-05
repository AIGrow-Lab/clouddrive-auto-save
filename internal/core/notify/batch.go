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
