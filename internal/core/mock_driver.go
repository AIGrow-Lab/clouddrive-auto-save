package core

import (
	"context"
	"fmt"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"time"
)

// MockDriver 模拟网盘驱动
type MockDriver struct {
	Files         []FileInfo // 用于 ListFiles
	ShareFiles    []FileInfo // 用于 ParseShare
	SaveLinkCalls int
	SavedFileIDs  []string
	TargetPaths   []string
}

func (m *MockDriver) GetInfo(ctx context.Context) (*db.Account, error) {
	return &db.Account{
		Platform:      "mock",
		Nickname:      "MockUser",
		Status:        1,
		CapacityUsed:  500 * 1024 * 1024 * 1024,
		CapacityTotal: 1024 * 1024 * 1024 * 1024,
		VipName:       "超级会员",
		LastCheck:     time.Now(),
	}, nil
}

func (m *MockDriver) Login(ctx context.Context) error {
	return nil
}

func (m *MockDriver) ListFiles(ctx context.Context, parentID string) ([]FileInfo, error) {
	return m.Files, nil
}

func (m *MockDriver) CreateFolder(ctx context.Context, parentID, name string) (*FileInfo, error) {
	return &FileInfo{ID: "new_folder_id", Name: name, IsFolder: true}, nil
}

func (m *MockDriver) DeleteFile(ctx context.Context, fileID string) error {
	return nil
}

func (m *MockDriver) ParseShare(ctx context.Context, shareURL, extractCode string) ([]FileInfo, error) {
	if len(m.ShareFiles) > 0 {
		return m.ShareFiles, nil
	}
	// 提供一些默认的分享内容
	return []FileInfo{
		{ID: "f1", Name: "功夫熊猫4.mp4", Size: 1024 * 1024 * 1024, UpdateTime: time.Now()},
		{ID: "f2", Name: "readme.txt", Size: 1024, UpdateTime: time.Now()},
	}, nil
}

func (m *MockDriver) SaveLink(ctx context.Context, shareURL, extractCode, targetPath string, fileIDs []string) error {
	m.SaveLinkCalls++
	m.SavedFileIDs = append(m.SavedFileIDs, fileIDs...)
	m.TargetPaths = append(m.TargetPaths, targetPath)
	return nil
}

func (m *MockDriver) RenameFile(ctx context.Context, fileID, newName string) error {
	return nil
}

func (m *MockDriver) SaveFileTo(ctx context.Context, fileID, targetPath string) error {
	return nil
}

func (m *MockDriver) PrepareTargetPath(ctx context.Context, path string) (string, error) {
	return "target_root_id", nil
}

// SetupE2EMock 开启 E2E 模式下的全局 Mock
func SetupE2EMock() {
	RegisterDriver("mock_139", func(account *db.Account) CloudDrive {
		return &MockDriver{}
	})
	RegisterDriver("mock_quark", func(account *db.Account) CloudDrive {
		return &MockDriver{}
	})
	fmt.Println("E2E Mock Drivers Registered")
}
