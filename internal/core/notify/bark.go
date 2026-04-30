package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// BarkPayload Bark 推送请求载荷
type BarkPayload struct {
	DeviceKey string `json:"device_key"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Level     string `json:"level,omitempty"`
	Badge     int    `json:"badge,omitempty"`
	Sound     string `json:"sound,omitempty"`
	Icon      string `json:"icon,omitempty"`
	Group     string `json:"group,omitempty"`
	URL       string `json:"url,omitempty"`
	IsArchive int    `json:"isArchive,omitempty"`
}

// SendBark 发送 Bark 推送
func SendBark(title, body, level, sound, icon, archive string) error {
	var enabledSetting, serverSetting, keySetting db.Setting

	// 获取配置
	db.DB.Where("key = ?", "bark_enabled").First(&enabledSetting)
	if enabledSetting.Value != "true" {
		return nil
	}

	db.DB.Where("key = ?", "bark_server").First(&serverSetting)
	db.DB.Where("key = ?", "bark_device_key").First(&keySetting)

	server := serverSetting.Value
	if server == "" {
		server = "https://api.day.app"
	}
	key := keySetting.Value
	if key == "" {
		return fmt.Errorf("bark device key is empty")
	}

	return SendBarkDirect(server, key, title, body, level, sound, icon, archive)
}

// SendBarkDirect 直接通过提供的服务器和 Key 发送推送（不检查开关）
func SendBarkDirect(server, key, title, body, level, sound, icon, archive string) error {
	if server == "" {
		server = "https://api.day.app"
	}
	if key == "" {
		return fmt.Errorf("bark device key is empty")
	}

	// 处理默认铃声
	if sound == "default" {
		sound = ""
	}

	isArchive := 1
	if archive == "false" {
		isArchive = 0
	}

	payload := BarkPayload{
		DeviceKey: key,
		Title:     title,
		Body:      body,
		Level:     level,
		Sound:     sound,
		Icon:      icon,
		Group:     "UCAS",
		IsArchive: isArchive,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 构造推送 URL
	pushURL := fmt.Sprintf("%s/push", server)
	req, err := http.NewRequest("POST", pushURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bark push failed with status: %d", resp.StatusCode)
	}

	slog.Debug("Bark 推送成功", "title", title)
	return nil
}

// SendTaskNotification 发送任务完成通知
func SendTaskNotification(taskName string, status string, message string, files []string, duration time.Duration) {
	var enabledSetting, serverSetting, keySetting, iconSetting, archiveSetting db.Setting

	// 同步读取所有配置，确保 goroutine 中不访问 db
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

	title := fmt.Sprintf("✅ 转存任务完成: %s", taskName)
	level := "active"
	sound := "birdsong.caf" // 成功时默认

	var levelSetting, soundSetting db.Setting
	if status == "failed" {
		title = fmt.Sprintf("❌ 转存任务失败: %s", taskName)
		level = "timeSensitive" // 失败时默认
		sound = "alarm.caf"     // 失败时默认

		db.DB.Where("key = ?", "bark_failure_level").First(&levelSetting)
		if levelSetting.Value != "" {
			level = levelSetting.Value
		}
		db.DB.Where("key = ?", "bark_failure_sound").First(&soundSetting)
		if soundSetting.Value != "" && soundSetting.Value != "default" {
			sound = soundSetting.Value
		} else if soundSetting.Value == "default" {
			sound = ""
		}
	} else {
		db.DB.Where("key = ?", "bark_success_level").First(&levelSetting)
		if levelSetting.Value != "" {
			level = levelSetting.Value
		}
		db.DB.Where("key = ?", "bark_success_sound").First(&soundSetting)
		if soundSetting.Value != "" && soundSetting.Value != "default" {
			sound = soundSetting.Value
		} else if soundSetting.Value == "default" {
			sound = ""
		}
	}

	icon := iconSetting.Value
	archive := archiveSetting.Value

	body := fmt.Sprintf("%s\n执行耗时: %s", message, duration.Round(time.Second))
	if len(files) > 0 {
		fileList := ""
		maxFiles := 10
		for i, f := range files {
			if i >= maxFiles {
				fileList += fmt.Sprintf("\n... 等共 %d 个文件", len(files))
				break
			}
			fileList += fmt.Sprintf("\n- %s", f)
		}
		body = fmt.Sprintf("%s\n\n转存文件列表:%s", body, fileList)
	}

	go func() {
		if err := SendBarkDirect(server, key, title, body, level, sound, icon, archive); err != nil {
			slog.Error("发送 Bark 通知失败", "err", err)
		}
	}()
}
