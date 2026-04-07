package main

import (
	"log"
	"os"
	"github.com/zcq/clouddrive-auto-save/internal/api"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

func main() {
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

	// 3. 启动 API 服务
	log.Println("Starting API server on :8080...")
	r := api.InitRouter(wm)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}
