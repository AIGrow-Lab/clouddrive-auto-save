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
	Group     string `json:"group,omitempty"`
	Icon      string `json:"icon,omitempty"`
}

// SendBark 发送 Bark 推送
func SendBark(title, body string) error {
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

	payload := BarkPayload{
		DeviceKey: key,
		Title:     title,
		Body:      body,
		Group:     "UCAS",
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
func SendTaskNotification(taskName string, status string, message string, files []string) {
	title := fmt.Sprintf("转存任务完成: %s", taskName)
	if status == "failed" {
		title = fmt.Sprintf("转存任务失败: %s", taskName)
	}

	body := message
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
		body = fmt.Sprintf("%s\n\n转存文件列表:%s", message, fileList)
	}

	go func() {
		if err := SendBark(title, body); err != nil {
			slog.Error("发送 Bark 通知失败", "err", err)
		}
	}()
}
