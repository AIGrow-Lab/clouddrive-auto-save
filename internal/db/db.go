package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

// Account 代表云盘账号 (139 或 Quark)
type Account struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Platform    string `gorm:"size:20;not null" json:"platform"` // "139" 或 "quark"
	Nickname    string `gorm:"size:100" json:"nickname"`
	AccountName string `gorm:"size:100" json:"account_name"`
	Cookie      string `gorm:"type:text" json:"cookie"`
	AuthToken   string `gorm:"type:text" json:"auth_token"` // 主要用于 139 的 Authorization
	Status      int    `gorm:"default:1" json:"status"`      // 1: 正常, 0: 失效
	CapacityUsed  int64  `json:"capacity_used"`
	CapacityTotal int64  `json:"capacity_total"`
	VipName       string `gorm:"size:50" json:"vip_name"`
	LastCheck   time.Time `json:"last_check"`
}

// Task 代表自动转存任务
type Task struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	AccountID   uint    `json:"account_id"`
	Account     Account `gorm:"foreignKey:AccountID" json:"account"`
	Name        string  `gorm:"size:255;not null" json:"name"`
	ShareURL    string  `gorm:"size:500;not null" json:"share_url"`
	ExtractCode string  `gorm:"size:50" json:"extract_code"`
	SavePath    string  `gorm:"size:500;not null" json:"save_path"`
	
	// 整理规则
	Pattern     string `gorm:"size:500" json:"pattern"`      // 正则表达式
	Replacement string `gorm:"size:500" json:"replacement"`  // 替换规则/魔法变量
	Filter      string `gorm:"size:500" json:"filter"`       // 过滤规则
	
	Cron        string `gorm:"size:100" json:"cron"`         // 任务独立 Cron (可选)
	StartDate   *time.Time `json:"start_date"`               // 起始日期过滤 (可选)
	StartFileID string     `gorm:"size:255" json:"start_file_id"` // 起始文件 ID (可选)
	LastRun     time.Time  `json:"last_run"`

	NextRun     time.Time `json:"next_run"`
	Status      string `gorm:"size:20;default:'pending'" json:"status"` // pending, running, success, failed
	Message     string `gorm:"type:text" json:"message"`     // 最后运行的错误信息或统计
}

// CommonFolder 常用目录
type CommonFolder struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	AccountID uint    `json:"account_id"`
	Path      string  `gorm:"size:500;not null" json:"path"`
	Name      string  `gorm:"size:255" json:"name"`
}

// Setting 系统全局设置
type Setting struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

var DB *gorm.DB

func InitDB(path string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移模型
	return DB.AutoMigrate(&Account{}, &Task{}, &CommonFolder{}, &Setting{})
}
