package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/zcq/clouddrive-auto-save/internal/api"
	"github.com/zcq/clouddrive-auto-save/internal/core/scheduler"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
)

func main() {
	// 0. 初始化日志系统
	logLevelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	minLevel := slog.LevelInfo
	switch logLevelStr {
	case "DEBUG":
		minLevel = slog.LevelDebug
	case "WARN":
		minLevel = slog.LevelWarn
	case "ERROR":
		minLevel = slog.LevelError
	}
	utils.InitLogger(minLevel, os.Stdout)
	slog.Info("Logging system initialized", "level", minLevel.String())

	// 1. 初始化数据库
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data.db"
	}
	slog.Info("Initializing database...", "path", dbPath)
	if err := db.InitDB(dbPath); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// 1.5 清理异常中断的任务（重置卡在 running 状态的任务）
	db.DB.Model(&db.Task{}).Where("status = ?", "running").Updates(map[string]interface{}{
		"status":  "pending",
		"message": "服务重启，已重置执行状态",
	})

	// 2. 启动任务管理器 (并发数为 3)
	slog.Info("Starting worker manager...", "workers", 3)
	wm := worker.NewManager(3, db.DB)
	wm.Start()
	defer wm.Stop()

	// 2.5 启动定时调度器
	scheduler.Init(wm)
	scheduler.Global.Start()
	defer scheduler.Global.Stop()

	// 加载全局调度设置
	var enabledSetting db.Setting
	var cronSetting db.Setting
	db.DB.Where("key = ?", "global_schedule_enabled").First(&enabledSetting)
	db.DB.Where("key = ?", "global_schedule_cron").First(&cronSetting)
	scheduler.Global.UpdateGlobalSchedule(cronSetting.Value, enabledSetting.Value == "true")

	// 加载所有任务的调度
	var tasks []db.Task
	db.DB.Find(&tasks)
	for _, t := range tasks {
		scheduler.Global.UpdateTask(t.ID, t.ScheduleMode, t.Cron)
	}

	// 3. 启动 API 服务
	slog.Info("Starting API server on 127.0.0.1:8080...")
	r := api.InitRouter(wm)
	if err := r.Run("127.0.0.1:8080"); err != nil {
		slog.Error("Failed to start API server", "error", err)
		os.Exit(1)
	}
}
