package worker

import (
	"context"

	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// MockDriver 模拟网盘驱动
type MockDriver struct {
	Files []core.FileInfo
}

func (m *MockDriver) GetInfo(ctx context.Context) (*db.Account, error) {
	return &db.Account{}, nil
}

func (m *MockDriver) Login(ctx context.Context) error {
	return nil
}

func (m *MockDriver) ListFiles(ctx context.Context, parentID string) ([]core.FileInfo, error) {
	return m.Files, nil
}

func (m *MockDriver) CreateFolder(ctx context.Context, parentID, name string) (*core.FileInfo, error) {
	return &core.FileInfo{ID: "new_folder_id", Name: name, IsFolder: true}, nil
}

func (m *MockDriver) DeleteFile(ctx context.Context, fileID string) error {
	return nil
}

func (m *MockDriver) ParseShare(ctx context.Context, shareURL, extractCode string) ([]core.FileInfo, error) {
	return m.Files, nil
}

func (m *MockDriver) SaveLink(ctx context.Context, shareURL, extractCode, targetPath string, fileIDs []string) error {
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
