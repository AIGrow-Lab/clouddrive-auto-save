package worker

import (
	"testing"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	testDB.AutoMigrate(&db.Account{}, &db.Task{})
	return testDB
}

func TestManager_Execute(t *testing.T) {
	testDB := setupTestDB(t)
	m := NewManager(1, testDB)

	// 注册 Mock 驱动
	core.RegisterDriver("mock", func(account *db.Account) core.CloudDrive {
		return &MockDriver{
			Files: []core.FileInfo{
				{ID: "f1", Name: "file1.mp4", UpdateTime: time.Now()},
				{ID: "f2", Name: "file2.mp4", UpdateTime: time.Now()},
			},
		}
	})

	account := db.Account{Platform: "mock", Nickname: "TestUser"}
	testDB.Create(&account)

	task := db.Task{
		AccountID: account.ID,
		Account:   account,
		Name:      "TestTask",
		ShareURL:  "http://share.com/1",
		SavePath:  "/test",
		Status:    "pending",
	}
	testDB.Create(&task)

	// 执行任务
	m.execute(&task)

	// 验证结果
	var updatedTask db.Task
	testDB.First(&updatedTask, task.ID)
	if updatedTask.Status != "success" {
		t.Errorf("expected task status success, got %s", updatedTask.Status)
	}
	if updatedTask.Percent != 100 {
		t.Errorf("expected task percent 100, got %d", updatedTask.Percent)
	}
}

func TestManager_Execute_SkipExisting(t *testing.T) {
	testDB := setupTestDB(t)
	m := NewManager(1, testDB)

	core.RegisterDriver("mock_skip", func(account *db.Account) core.CloudDrive {
		return &MockDriver{
			Files: []core.FileInfo{
				{ID: "f1", Name: "file1.mp4", UpdateTime: time.Now()},
			},
		}
	})

	account := db.Account{Platform: "mock_skip", Nickname: "TestUser"}
	testDB.Create(&account)

	task := db.Task{
		AccountID: account.ID,
		Account:   account,
		Name:      "TestTask",
		ShareURL:  "http://share.com/1",
		SavePath:  "/test",
		Status:    "pending",
	}
	testDB.Create(&task)

	m.execute(&task)

	var updatedTask db.Task
	testDB.First(&updatedTask, task.ID)
	if updatedTask.Status != "success" {
		t.Errorf("expected success, got %s", updatedTask.Status)
	}
}
