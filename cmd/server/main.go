package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/zcq/clouddrive-auto-save/internal/api"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/core/scheduler"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
)

// 版本信息，构建时通过 -ldflags 注入
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
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
	slog.Info("UCAS starting", "version", version, "commit", commit, "date", date, "logLevel", minLevel.String())

	// 1. 初始化数据库
	dbPath := os.Getenv("DB_PATH")
	isE2E := os.Getenv("E2E_TEST_MODE") == "true"
	if isE2E {
		dbPath = "file::memory:?cache=shared"
		slog.Info("Running in E2E TEST MODE (using memory database)")
		// 开启 HTTP 层 Mock 拦截，让系统走真实驱动逻辑进行 JSON 解析测试
		core.SetupE2EHTTPMock()
	} else if dbPath == "" {
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

	// 设置默认值：每天凌晨
	if cronSetting.Value == "" {
		cronSetting.Key = "global_schedule_cron"
		cronSetting.Value = "0 0 0 * * *"
		db.DB.Save(&cronSetting)
		slog.Info("Initialized default global schedule cron: 0 0 0 * * *")
	}

	scheduler.Global.UpdateGlobalSchedule(cronSetting.Value, enabledSetting.Value == "true")

	// 加载所有任务的调度
	var tasks []db.Task
	db.DB.Find(&tasks)
	for _, t := range tasks {
		scheduler.Global.UpdateTask(t.ID, t.ScheduleMode, t.Cron)
	}

	// 3. 启动 API 服务
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "0.0.0.0:8080"
	}
	slog.Info("Starting API server", "addr", listenAddr)
	r := api.InitRouter(wm, version, commit, date)
	if err := r.Run(listenAddr); err != nil {
		slog.Error("Failed to start API server", "error", err)
		os.Exit(1)
	}
}
