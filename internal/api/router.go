package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/core/scheduler"
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
		api.POST("/tasks/:id/dismiss", dismissTaskAPI)
		api.POST("/tasks/preview", previewTask)
		api.POST("/tasks/parse_share", parseShareLinkInfo)

		api.GET("/dashboard/stats", getDashboardStats)
		api.GET("/dashboard/logs", streamLogs)
		api.GET("/dashboard/logs/recent", getRecentLogs)
		api.DELETE("/dashboard/logs/recent", clearRecentLogs)

		api.GET("/settings/schedule", getScheduleSettings)
		api.POST("/settings/schedule", updateScheduleSettings)
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
		now := time.Now()
		account.Status = 0
		account.LastCheck = now
		db.DB.Model(account).Updates(map[string]interface{}{
			"status":     0,
			"last_check": now,
		})
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

	// 校验 Cron 表达式
	if task.ScheduleMode == "custom" {
		if err := scheduler.ValidateCron(task.Cron); err != nil {
			c.PureJSON(http.StatusBadRequest, gin.H{"error": "Cron 表达式格式错误: " + err.Error()})
			return
		}
	}

	log.Printf("[API] 创建任务: %s", task.Name)
	db.DB.Create(&task)

	// 推送实时事件
	utils.BroadcastTaskUpdate(&task)
	utils.BroadcastStatsUpdate()

	// 注册定时任务
	scheduler.Global.UpdateTask(task.ID, task.ScheduleMode, task.Cron)

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

	// 校验 Cron 表达式
	if task.ScheduleMode == "custom" {
		if err := scheduler.ValidateCron(task.Cron); err != nil {
			c.PureJSON(http.StatusBadRequest, gin.H{"error": "Cron 表达式格式错误: " + err.Error()})
			return
		}
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
		"cron":            task.Cron,
		"schedule_mode":   task.ScheduleMode,
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

	// 推送更新事件
	utils.BroadcastTaskUpdate(&task)

	// 刷新调度器
	scheduler.Global.UpdateTask(task.ID, task.ScheduleMode, task.Cron)

	c.PureJSON(http.StatusOK, task)
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[API] 删除任务: ID=%s", id)

	idNum, _ := strconv.Atoi(id)
	scheduler.Global.RemoveTask(uint(idNum))

	db.DB.Delete(&db.Task{}, id)

	// 推送实时事件
	utils.BroadcastTaskDelete(uint(idNum))
	utils.BroadcastStatsUpdate()

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

	// 立即更新状态并推送
	task.Status = "running"
	task.Stage = "Started" // 重置 Dismissed 状态
	db.DB.Model(&task).Updates(map[string]interface{}{
		"status": "running",
		"stage":  "Started",
	})
	utils.BroadcastTaskUpdate(&task)
	utils.BroadcastStatsUpdate()

	WorkerManager.Submit(worker.Job{Task: &task})
	c.PureJSON(http.StatusOK, gin.H{"message": "task submitted to worker pool"})
}

func dismissTaskAPI(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Model(&db.Task{}).Where("id = ?", id).Update("stage", "Dismissed").Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.PureJSON(http.StatusOK, gin.H{"message": "task dismissed"})
}

func getDashboardStats(c *gin.Context) {
	// 获取全局调度开关状态
	var enabledSetting db.Setting
	db.DB.Where("key = ?", "global_schedule_enabled").First(&enabledSetting)
	globalEnabled := enabledSetting.Value == "true"

	var scheduledTasks int64
	if globalEnabled {
		// 全局开启：统计模式为 global 的任务 + 模式为 custom 且有 cron 的任务
		db.DB.Model(&db.Task{}).Where("schedule_mode = ? OR (schedule_mode = ? AND cron != '')", "global", "custom").Count(&scheduledTasks)
	} else {
		// 全局关闭：仅统计模式为 custom 且有 cron 的任务
		db.DB.Model(&db.Task{}).Where("schedule_mode = ? AND cron != ''", "custom").Count(&scheduledTasks)
	}

	var capacityUsed int64
	db.DB.Model(&db.Account{}).Where("status = 1").Select("SUM(capacity_used)").Scan(&capacityUsed)

	var todayCompleted int64
	db.DB.Model(&db.Task{}).Where("status = ? AND DATE(last_run) = DATE('now', 'localtime')", "success").Count(&todayCompleted)

	var activeAccounts int64
	db.DB.Model(&db.Account{}).Where("status = 1").Count(&activeAccounts)

	var runningTasksList []db.Task
	// 获取：运行中、15 秒内成功的任务、以及未被忽略（Dismissed）的失败任务
	db.DB.Where("status = ? OR (status = ? AND last_run >= ?) OR (status = ? AND stage != ?)",
		"running", "success", time.Now().Add(-15*time.Second), "failed", "Dismissed").Find(&runningTasksList)

	var recentTasks []db.Task
	db.DB.Order("last_run desc").Limit(5).Find(&recentTasks)

	c.PureJSON(http.StatusOK, gin.H{
		"scheduled_tasks":    scheduledTasks,
		"capacity_used":      capacityUsed,
		"today_completed":    todayCompleted,
		"active_accounts":    activeAccounts,
		"recent_activities":  recentTasks,
		"running_tasks_list": runningTasksList,
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

func getScheduleSettings(c *gin.Context) {
	var enabledSetting db.Setting
	var cronSetting db.Setting

	db.DB.Where("key = ?", "global_schedule_enabled").First(&enabledSetting)
	db.DB.Where("key = ?", "global_schedule_cron").First(&cronSetting)

	c.PureJSON(http.StatusOK, gin.H{
		"enabled": enabledSetting.Value == "true",
		"cron":    cronSetting.Value,
	})
}

func updateScheduleSettings(c *gin.Context) {
	var input struct {
		Enabled bool   `json:"enabled"`
		Cron    string `json:"cron"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 校验全局 Cron 表达式
	if input.Enabled {
		if err := scheduler.ValidateCron(input.Cron); err != nil {
			c.PureJSON(http.StatusBadRequest, gin.H{"error": "全局 Cron 表达式格式错误: " + err.Error()})
			return
		}
	}

	enabledStr := "false"
	if input.Enabled {
		enabledStr = "true"
	}

	// 使用 Updates 或 FirstOrCreate 确保 Key 存在
	db.DB.Save(&db.Setting{Key: "global_schedule_enabled", Value: enabledStr})
	db.DB.Save(&db.Setting{Key: "global_schedule_cron", Value: input.Cron})

	scheduler.Global.UpdateGlobalSchedule(input.Cron, input.Enabled)

	// 推送统计更新
	utils.BroadcastStatsUpdate()

	c.PureJSON(http.StatusOK, gin.H{"message": "settings updated"})
}
