package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/scheduler"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	testDB.AutoMigrate(&db.Account{}, &db.Task{}, &db.Setting{})
	db.DB = testDB // 设置全局 DB 供 handler 使用

	wm := worker.NewManager(1, testDB)
	scheduler.Init(wm)
	r := InitRouter(wm)
	return r, testDB
}

func TestListAccounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, testDB := setupTestRouter(t)

	// 插入测试数据
	testDB.Create(&db.Account{Platform: "139", Nickname: "User1"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/accounts", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var accounts []db.Account
	err := json.Unmarshal(w.Body.Bytes(), &accounts)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(accounts))
	}
	if accounts[0].Nickname != "User1" {
		t.Errorf("expected nickname User1, got %s", accounts[0].Nickname)
	}
}

func TestCreateAccountWithWhitespace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, testDB := setupTestRouter(t)

	// 构造带有内部换行符、首尾换行符和空格的请求数据
	accountData := map[string]interface{}{
		"platform":     "139",
		"account_name": "TestUser",
		"cookie":       "  my-\nsecret\r-cookie\n\n ",
		"auth_token":   "\nBearer\n secret-token  ",
	}
	body, _ := json.Marshal(accountData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// 从数据库验证数据是否已被清理
	var savedAccount db.Account
	testDB.First(&savedAccount, "account_name = ?", "TestUser")

	if savedAccount.Cookie != "my-secret-cookie" {
		t.Errorf("expected trimmed and cleaned cookie 'my-secret-cookie', got '%s'", savedAccount.Cookie)
	}
	if savedAccount.AuthToken != "Bearer secret-token" {
		t.Errorf("expected trimmed and cleaned auth_token 'Bearer secret-token', got '%s'", savedAccount.AuthToken)
	}
}

func TestUpdateAccountWithWhitespace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, testDB := setupTestRouter(t)

	// 先创建一个正常账号
	account := db.Account{
		Platform:    "quark",
		AccountName: "UpdateUser",
		Cookie:      "old-cookie",
	}
	testDB.Create(&account)

	// 构造更新数据，带有多余空白
	updateData := map[string]interface{}{
		"platform":     "quark",
		"account_name": "UpdateUser",
		"cookie":       "  new-cookie\n ",
		"auth_token":   "  new-token  ",
	}
	body, _ := json.Marshal(updateData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/accounts/"+strconv.Itoa(int(account.ID)), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// 验证数据库
	var updatedAccount db.Account
	testDB.First(&updatedAccount, account.ID)

	if updatedAccount.Cookie != "new-cookie" {
		t.Errorf("expected trimmed cookie 'new-cookie', got '%s'", updatedAccount.Cookie)
	}
	if updatedAccount.AuthToken != "new-token" {
		t.Errorf("expected trimmed auth_token 'new-token', got '%s'", updatedAccount.AuthToken)
	}
}

func TestUpdateGlobalSettings_InvalidCron(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, testDB := setupTestRouter(t)

	// 先开启全局调度
	testDB.Save(&db.Setting{Key: "global_schedule_enabled", Value: "true"})

	// 构造非法的 Cron 表达式请求
	settingsData := map[string]string{
		"global_schedule_cron": "invalid-cron",
	}
	body, _ := json.Marshal(settingsData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/settings/global", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !strings.Contains(resp["error"], "Cron 表达式格式错误") {
		t.Errorf("expected cron error message, got: %s", resp["error"])
	}
}
