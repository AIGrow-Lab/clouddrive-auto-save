// internal/core/openlist/client_test.go
package openlist

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_StartScan_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admin/scan/start" {
			t.Errorf("期望路径 /api/admin/scan/start，实际 %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "test-token" {
			t.Errorf("期望 Authorization test-token，实际 %s", r.Header.Get("Authorization"))
		}
		if r.Method != http.MethodPost {
			t.Errorf("期望 POST 方法，实际 %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.StartScan(context.Background())
	if err != nil {
		t.Fatalf("期望成功，实际错误: %v", err)
	}
}

func TestClient_StartScan_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.StartScan(context.Background())
	if err == nil {
		t.Fatal("期望错误，实际成功")
	}
	if !contains(err.Error(), "500") {
		t.Errorf("错误信息应包含状态码 500，实际: %v", err)
	}
}

func TestClient_StartScan_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	client.httpClient.Timeout = 50 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.StartScan(ctx)
	if err == nil {
		t.Fatal("期望超时错误，实际成功")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
