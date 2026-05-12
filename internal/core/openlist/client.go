// internal/core/openlist/client.go
package openlist

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// scanResponse OpenList API 响应结构
type scanResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

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
	url := c.baseURL + "/api/admin/scan/start"
	slog.Debug("OpenList 发送扫描请求", "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
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

	body, _ := io.ReadAll(resp.Body)
	slog.Debug("OpenList 扫描响应",
		"status", resp.StatusCode,
		"body", string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	// 检查业务错误码（OpenList 返回 HTTP 200 但可能包含业务错误）
	// code=200 或 code=0 表示成功
	var result scanResponse
	if err := json.Unmarshal(body, &result); err == nil && result.Code != 0 && result.Code != 200 {
		return fmt.Errorf("API 业务错误 %d: %s", result.Code, result.Message)
	}

	return nil
}
