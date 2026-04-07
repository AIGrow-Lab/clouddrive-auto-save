package core

import (
	"context"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// FileInfo 代表云盘中的文件或文件夹信息
type FileInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ParentID  string `json:"parent_id"`
	IsFolder  bool   `json:"is_folder"`
	Size      int64  `json:"size"`
	UpdatedAt string `json:"updated_at"`
}

// CloudDrive 定义了所有云盘必须实现的标准接口
type CloudDrive interface {
	// 账号相关
	GetInfo(ctx context.Context) (*db.Account, error)
	Login(ctx context.Context) error
	
	// 文件操作
	ListFiles(ctx context.Context, parentID string) ([]FileInfo, error)
	CreateFolder(ctx context.Context, name, parentID string) (string, error)
	DeleteFile(ctx context.Context, fileID string) error
	
	// 分享转存相关
	// SaveLink 将分享链接中的文件转存到指定目标目录
	SaveLink(ctx context.Context, shareURL, extractCode, targetPath string) error
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
