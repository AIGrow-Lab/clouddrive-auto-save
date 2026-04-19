package main

import (
	"io"
	"log"
	"os"

	"github.com/zcq/clouddrive-auto-save/internal/api"
	"github.com/zcq/clouddrive-auto-save/internal/core/scheduler"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
)

type BroadcasterWriter struct{}

func (w *BroadcasterWriter) Write(p []byte) (n int, err error) {
	utils.GlobalBroadcaster.Broadcast(string(p))
	return len(p), nil
}

func main() {
	// 0. 劫持日志输出，镜像到广播器
	log.SetOutput(io.MultiWriter(os.Stdout, &BroadcasterWriter{}))

	// 1. 初始化数据库

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data.db"
	}
	log.Printf("Initializing database at %s...", dbPath)
	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 1.5 清理异常中断的任务（重置卡在 running 状态的任务）
	db.DB.Model(&db.Task{}).Where("status = ?", "running").Updates(map[string]interface{}{
		"status":  "pending",
		"message": "服务重启，已重置执行状态",
	})

	// 2. 启动任务管理器 (并发数为 3)
	log.Println("Starting worker manager...")
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
	log.Println("Starting API server on 127.0.0.1:8080...")
	r := api.InitRouter(wm)
	if err := r.Run("127.0.0.1:8080"); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}
