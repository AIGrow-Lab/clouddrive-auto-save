package core

import (
	"context"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"time"
)

// FileInfo 代表云盘中的文件或文件夹信息
type FileInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"` // 全路径，用于部分 API
	ParentID   string    `json:"parent_id"`
	IsFolder   bool      `json:"is_folder"`
	Size       int64     `json:"size"`
	UpdatedAt  string    `json:"updated_at"`
	UpdateTime time.Time `json:"-"`          // 解析后的标准时间
	IsExisted  bool      `json:"is_existed"` // 标记是否已在目标目录存在
	NewName    string    `json:"new_name"`   // 重命名后的预览名
}

// CloudDrive 定义了所有云盘必须实现的标准接口
type CloudDrive interface {
	// 账号相关
	GetInfo(ctx context.Context) (*db.Account, error)
	Login(ctx context.Context) error

	// 文件操作
	ListFiles(ctx context.Context, parentID string) ([]FileInfo, error)
	CreateFolder(ctx context.Context, parentID, name string) (*FileInfo, error)
	DeleteFile(ctx context.Context, fileID string) error

	// 分享转存相关
	// ParseShare 解析分享链接，返回其中的文件列表
	ParseShare(ctx context.Context, shareURL, extractCode string) ([]FileInfo, error)
	// PrepareTargetPath 准备目标目录，返回目标目录的 ID。如果目录不存在则创建。
	PrepareTargetPath(ctx context.Context, path string) (string, error)
	// RenameFile 修改文件名或目录名
	RenameFile(ctx context.Context, fileID, newName string) error
	// SaveFileTo 将特定的分享文件保存到目标路径
	SaveFileTo(ctx context.Context, fileID, targetPath string) error
	// SaveLink 将分享链接中的文件转存到指定目标目录。如果 fileIDs 不为空，则仅转存指定的文件。
	SaveLink(ctx context.Context, shareURL, extractCode, targetPath string, fileIDs []string) error
}

// DriveFactory 用于根据平台创建对应的驱动实例
type DriveFactory func(account *db.Account) CloudDrive

var drivers = make(map[string]DriveFactory)

func RegisterDriver(platform string, factory DriveFactory) {
	drivers[platform] = factory
}

func GetDriver(account *db.Account) CloudDrive {
	if factory, ok := drivers[account.Platform]; ok {
		return factory(account)
	}
	return nil
}
