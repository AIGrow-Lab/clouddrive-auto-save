package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// MockDriver 已经在 worker 包中定义过了，但由于 api 和 worker 是平级的
// 我们需要在这里也定义一个 Mock 驱动或者使用 interface。
// 实际上 internal/core/drive.go 中有 RegisterDriver 机制。

type apiMockDriver struct {
	Files []core.FileInfo
}

func (m *apiMockDriver) GetInfo(ctx context.Context) (*db.Account, error) { return &db.Account{}, nil }
func (m *apiMockDriver) Login(ctx context.Context) error                  { return nil }
func (m *apiMockDriver) ListFiles(ctx context.Context, p string) ([]core.FileInfo, error) {
	return m.Files, nil
}
func (m *apiMockDriver) CreateFolder(ctx context.Context, p, n string) (*core.FileInfo, error) {
	return &core.FileInfo{ID: "dir1", Name: n}, nil
}
func (m *apiMockDriver) DeleteFile(ctx context.Context, id string) error { return nil }
func (m *apiMockDriver) ParseShare(ctx context.Context, u, p string) ([]core.FileInfo, error) {
	return m.Files, nil
}
func (m *apiMockDriver) SaveLink(ctx context.Context, u, p, t string, ids []string) error { return nil }
func (m *apiMockDriver) RenameFile(ctx context.Context, id, n string) error               { return nil }
func (m *apiMockDriver) SaveFileTo(ctx context.Context, id, t string) error               { return nil }
func (m *apiMockDriver) PrepareTargetPath(ctx context.Context, p string) (string, error) {
	return "root", nil
}

func TestTaskPreview_Contract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, testDB := setupTestRouter(t)

	// 注册 Mock 驱动并准备数据
	core.RegisterDriver("mock_api", func(account *db.Account) core.CloudDrive {
		return &apiMockDriver{
			Files: []core.FileInfo{
				{ID: "f1", Name: "功夫熊猫4.mp4", UpdateTime: time.Now()},
			},
		}
	})

	account := db.Account{Platform: "mock_api", Nickname: "TestUser"}
	testDB.Create(&account)

	// 1. 测试基础响应字段契约
	payload := map[string]interface{}{
		"account_id":  account.ID,
		"share_url":   "http://share.com/1",
		"pattern":     "功夫",
		"replacement": "KungFu",
		"name":        "TestTask",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/tasks/preview", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var results []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &results)

	if len(results) == 0 {
		t.Fatal("expected results, got 0")
	}

	res := results[0]
	// 强制要求以下字段存在且符合契约，防止 AI 乱改
	requiredFields := []string{"original_name", "new_name", "matched", "is_filtered"}
	for _, field := range requiredFields {
		if _, ok := res[field]; !ok {
			t.Errorf("contract violation: missing field %s in preview response", field)
		}
	}

	// 2. 验证正则不匹配时的逻辑
	payload["pattern"] = "不匹配的正则"
	body, _ = json.Marshal(payload)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/tasks/preview", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &results)
	res = results[0]
	if res["matched"].(bool) != false {
		t.Error("expected matched to be false for non-matching regex")
	}
	if res["new_name"].(string) != "功夫熊猫4.mp4" {
		t.Errorf("expected new_name to be original name when not matched, got %s", res["new_name"])
	}

	// 3. 验证正则为空时的逻辑
	payload["pattern"] = ""
	body, _ = json.Marshal(payload)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/tasks/preview", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &results)
	res = results[0]
	if res["matched"].(bool) != true {
		t.Error("expected matched to be true for empty regex")
	}
}
