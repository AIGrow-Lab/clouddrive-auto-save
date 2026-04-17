package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	_ "github.com/zcq/clouddrive-auto-save/internal/core/cloud139"
	_ "github.com/zcq/clouddrive-auto-save/internal/core/quark"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

var WorkerManager *worker.Manager

func InitRouter(wm *worker.Manager) *gin.Engine {
	WorkerManager = wm
	r := gin.Default()

	// 基础 API 路由组
	api := r.Group("/api")
	{
		api.GET("/accounts", listAccounts)
		api.POST("/accounts", createAccount)
		api.PUT("/accounts/:id", updateAccount)
		api.DELETE("/accounts/:id", deleteAccount)
		api.POST("/accounts/:id/check", checkAccount)
		api.GET("/accounts/:id/folders", getAccountFolders)
		api.POST("/accounts/:id/folders", createAccountFolder)

		api.GET("/tasks", listTasks)
		api.POST("/tasks", createTask)
		api.PUT("/tasks/:id", updateTask)
		api.DELETE("/tasks/:id", deleteTask)
		api.POST("/tasks/:id/run", runTask)
		api.POST("/tasks/preview", previewTask)
		api.POST("/tasks/parse_share", parseShareLinkInfo)

		api.GET("/dashboard/stats", getDashboardStats)
		api.GET("/dashboard/logs", streamLogs)
		api.GET("/dashboard/logs/recent", getRecentLogs)
		api.DELETE("/dashboard/logs/recent", clearRecentLogs)
	}

	// 静态资源处理
	r.NoRoute(func(c *gin.Context) {
		http.FileServer(GetStaticFS()).ServeHTTP(c.Writer, c.Request)
	})

	return r
}

func clearRecentLogs(c *gin.Context) {
	utils.GlobalBroadcaster.ClearRecent()
	c.PureJSON(http.StatusOK, gin.H{"message": "logs cleared"})
}

func performAccountCheck(account *db.Account, ctx context.Context) error {
	driver := core.GetDriver(account)
	if driver == nil {
		return fmt.Errorf("driver not found for platform: %s", account.Platform)
	}

	updatedAccount, err := driver.GetInfo(ctx)
	if err != nil {
		account.Status = 0
		db.DB.Model(account).Update("status", 0)
		return err
	}

	err = db.DB.Model(account).Updates(map[string]interface{}{
		"nickname":       updatedAccount.Nickname,
		"account_name":   updatedAccount.AccountName,
		"status":         1,
		"capacity_used":  updatedAccount.CapacityUsed,
		"capacity_total": updatedAccount.CapacityTotal,
		"vip_name":       updatedAccount.VipName,
		"last_check":     time.Now(),
	}).Error
	if err != nil {
		return err
	}

	// 重新加载完整信息
	return db.DB.First(account, account.ID).Error
}

func checkAccount(c *gin.Context) {
	id := c.Param("id")
	var account db.Account
	if err := db.DB.First(&account, id).Error; err != nil {
		log.Printf("[API] 账号校验失败: 未找到 ID=%s", id)
		c.PureJSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	if err := performAccountCheck(&account, c.Request.Context()); err != nil {
		log.Printf("[API] 账号校验失败: %v", err)
		c.PureJSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "account": account})
		return
	}

	c.PureJSON(http.StatusOK, account)
}

func listAccounts(c *gin.Context) {
	var accounts []db.Account
	db.DB.Find(&accounts)
	c.PureJSON(http.StatusOK, accounts)
}

func createAccount(c *gin.Context) {
	var account db.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[API] 添加账号: %s (%s)", account.AccountName, account.Platform)
	if err := db.DB.Create(&account).Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	// 自动执行一次校验
	if err := performAccountCheck(&account, c.Request.Context()); err != nil {
		log.Printf("[API] 账号自动校验失败: %v", err)
	}

	c.PureJSON(http.StatusOK, account)
}

func updateAccount(c *gin.Context) {
	id := c.Param("id")
	var account db.Account
	if err := db.DB.First(&account, id).Error; err != nil {
		c.PureJSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	if err := c.ShouldBindJSON(&account); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[API] 更新账号: %s", account.AccountName)
	if err := db.DB.Save(&account).Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "failed to update account"})
		return
	}

	// 自动执行一次校验
	if err := performAccountCheck(&account, c.Request.Context()); err != nil {
		log.Printf("[API] 账号自动校验失败: %v", err)
	}

	c.PureJSON(http.StatusOK, account)
}

func deleteAccount(c *gin.Context) {
	id := c.Param("id")

	// 检查是否有关联任务
	var count int64
	db.DB.Model(&db.Task{}).Where("account_id = ?", id).Count(&count)
	if count > 0 {
		log.Printf("[API] 尝试删除账号失败: ID=%s, 存在 %d 个关联任务", id, count)
		c.PureJSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("该账号有关联的 %d 个任务，请先删除关联任务", count)})
		return
	}

	log.Printf("[API] 删除账号: ID=%s", id)
	db.DB.Delete(&db.Account{}, id)
	c.PureJSON(http.StatusOK, gin.H{"message": "deleted"})
}

func listTasks(c *gin.Context) {
	var tasks []db.Task
	db.DB.Preload("Account").Find(&tasks)
	c.PureJSON(http.StatusOK, tasks)
}

func createTask(c *gin.Context) {
	var task db.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[API] 创建任务: %s", task.Name)
	db.DB.Create(&task)
	c.PureJSON(http.StatusOK, task)
}

func updateTask(c *gin.Context) {
	id := c.Param("id")
	var task db.Task
	if err := db.DB.First(&task, id).Error; err != nil {
		c.PureJSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// 记录更新前的关键参数，用于判断是否需要重置状态
	oldURL := task.ShareURL
	oldCode := task.ExtractCode

	if err := c.ShouldBindJSON(&task); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[API] 更新任务: %s", task.Name)

	updateData := map[string]interface{}{
		"name":            task.Name,
		"account_id":      task.AccountID,
		"share_url":       task.ShareURL,
		"extract_code":    task.ExtractCode,
		"save_path":       task.SavePath,
		"pattern":         task.Pattern,
		"replacement":     task.Replacement,
		"start_file_id":   task.StartFileID,
		"start_file_name": task.StartFileName,
	}

	// 仅当分享链接或提取码发生变动时，才重置状态以解除 [Fatal] 封锁
	if task.ShareURL != oldURL || task.ExtractCode != oldCode {
		log.Printf("[API] 检测到关键参数变更，自动重置任务状态: %s", task.Name)
		updateData["status"] = "pending"
		updateData["message"] = ""
	}

	if err := db.DB.Model(&task).Updates(updateData).Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.PureJSON(http.StatusOK, task)
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[API] 删除任务: ID=%s", id)
	db.DB.Delete(&db.Task{}, id)
	c.PureJSON(http.StatusOK, gin.H{"message": "task deleted"})
}

func runTask(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	log.Printf("[API] 请求运行任务: ID=%d", id)

	var task db.Task
	if err := db.DB.Preload("Account").First(&task, id).Error; err != nil {
		c.PureJSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	if task.Status == "running" {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "task is already running"})
		return
	}

	WorkerManager.Submit(worker.Job{Task: &task})
	c.PureJSON(http.StatusOK, gin.H{"message": "task submitted to worker pool"})
}

func getDashboardStats(c *gin.Context) {
	var runningTasks int64
	db.DB.Model(&db.Task{}).Where("status = ?", "running").Count(&runningTasks)

	var capacityUsed int64
	db.DB.Model(&db.Account{}).Where("status = 1").Select("SUM(capacity_used)").Scan(&capacityUsed)

	var todayCompleted int64
	db.DB.Model(&db.Task{}).Where("status = ? AND DATE(last_run) = DATE('now', 'localtime')", "success").Count(&todayCompleted)

	var activeAccounts int64
	db.DB.Model(&db.Account{}).Where("status = 1").Count(&activeAccounts)

	var recentTasks []db.Task
	db.DB.Order("last_run desc").Limit(5).Find(&recentTasks)

	c.PureJSON(http.StatusOK, gin.H{
		"running_tasks":     runningTasks,
		"capacity_used":     capacityUsed,
		"today_completed":   todayCompleted,
		"active_accounts":   activeAccounts,
		"recent_activities": recentTasks,
	})
}

func streamLogs(c *gin.Context) {
	clientChan := utils.GlobalBroadcaster.Subscribe()
	defer utils.GlobalBroadcaster.Unsubscribe(clientChan)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case msg, ok := <-clientChan:
			if !ok {
				return false
			}
			c.SSEvent("message", msg)
			return true
		case <-heartbeat.C:
			c.SSEvent("heartbeat", "keep-alive")
			return true
		}
	})
}

func getRecentLogs(c *gin.Context) {
	logs := utils.GlobalBroadcaster.GetRecent()
	c.PureJSON(http.StatusOK, logs)
}
