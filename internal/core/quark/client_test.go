package quark

import (
	"context"
	"strings"
	"testing"

	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

func TestParseShare_EmptyList(t *testing.T) {
	// 启用全局 Mock
	core.SetupE2EHTTPMock()
	defer core.ResetMockState()

	// 构造测试账号
	account := &db.Account{
		Platform: "quark",
		Cookie:   "mock_normal",
	}

	client := NewQuark(account)

	// mock_empty: 会触发我们在 mock_http 中预设的返回空列表的响应
	shareURL := "https://pan.quark.cn/s/mock_empty"
	extractCode := ""

	files, err := client.ParseShare(context.Background(), shareURL, extractCode)

	if err == nil {
		t.Fatalf("expected error for empty share list, got nil. files: %v", files)
	}

	if !strings.Contains(err.Error(), "[Fatal]") {
		t.Errorf("expected fatal error, got: %v", err)
	}

	if !strings.Contains(err.Error(), "为空") {
		t.Errorf("expected error message about empty list, got: %v", err)
	}
}
