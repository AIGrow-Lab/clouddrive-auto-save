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

	// 2. 启动任务管理器 (并发数为 3)
	log.Println("Starting worker manager...")
	wm := worker.NewManager(3)
	wm.Start()
	defer wm.Stop()

	// 2.5 启动定时调度器
	scheduler.Init(wm)
	scheduler.Global.Start()
	defer scheduler.Global.Stop()

	// 加载现有定时任务
	var tasks []db.Task
	db.DB.Find(&tasks)
	for _, t := range tasks {
		if t.Cron != "" {
			scheduler.Global.UpdateTask(t.ID, t.Cron)
		}
	}

	// 3. 启动 API 服务
	log.Println("Starting API server on 127.0.0.1:8080...")
	r := api.InitRouter(wm)
	if err := r.Run("127.0.0.1:8080"); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}
