package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	_ "github.com/zcq/clouddrive-auto-save/internal/core/cloud139"
	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
	_ "github.com/zcq/clouddrive-auto-save/internal/core/quark"
	"github.com/zcq/clouddrive-auto-save/internal/core/scheduler"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
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
		api.POST("/tasks/run_all", runAllTasks)
		api.POST("/tasks/:id/dismiss", dismissTaskAPI)
		api.POST("/tasks/preview", previewTask)
		api.POST("/tasks/parse_share", parseShareLinkInfo)

		api.GET("/dashboard/stats", getDashboardStats)
		api.GET("/dashboard/logs", streamLogs)
		api.GET("/dashboard/logs/recent", getRecentLogs)
		api.DELETE("/dashboard/logs/recent", clearRecentLogs)

		api.GET("/settings/schedule", getScheduleSettings)
		api.POST("/settings/schedule", updateScheduleSettings)
		api.GET("/settings/global", getGlobalSettings)
		api.POST("/settings/global", updateGlobalSettings)
		api.POST("/settings/test_bark", testBarkNotification)
	}

	// 静态资源处理
	staticFS := GetStaticFS()
	fileServer := http.FileServer(staticFS)

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// 1. 如果是 API 请求，直接返回 404
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
			return
		}

		// 2. 如果是请求具体的文件（带点 .），尝试通过 FileServer 处理
		if strings.Contains(path, ".") {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// 3. 对于页面路由（如 /tasks 或 /），手动返回 index.html 内容
		// 使用 Open + ReadAll 绕过 http.ServeFile 的自动重定向逻辑
		f, err := staticFS.Open("index.html")
		if err != nil {
			slog.Error("无法打开 index.html", "error", err)
			c.String(http.StatusNotFound, "Frontend assets not found")
			return
		}
		defer f.Close()

		content, err := io.ReadAll(f)
		if err != nil {
			slog.Error("读取 index.html 失败", "error", err)
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	})

	return r
}

func clearRecentLogs(c *gin.Context) {
	utils.GlobalBroadcaster.ClearRecent()

	// 同时清理数据库中所有任务的最后运行消息和阶段
	// 增加 Where("1 = 1") 以绕过 GORM 的 AllowGlobalUpdate 限制
	db.DB.Model(&db.Task{}).Where("1 = 1").Updates(map[string]interface{}{
		"message": "",
		"stage":   "",
	})

	// 通知前端刷新
	utils.BroadcastStatsUpdate()

	c.PureJSON(http.StatusOK, gin.H{"message": "logs and task summaries cleared"})
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
		slog.Error("账号校验失败: 未找到", "account_id", id)
		c.PureJSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	if err := performAccountCheck(&account, c.Request.Context()); err != nil {
		slog.Error("账号校验失败", "account_id", id, "error", err)
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

	// 清理凭据中的空白字符和所有换行符（防止 HTTP Header 报错）
	account.Cookie = strings.ReplaceAll(strings.ReplaceAll(account.Cookie, "\n", ""), "\r", "")
	account.Cookie = strings.TrimSpace(account.Cookie)
	account.AuthToken = strings.ReplaceAll(strings.ReplaceAll(account.AuthToken, "\n", ""), "\r", "")
	account.AuthToken = strings.TrimSpace(account.AuthToken)

	slog.Info("添加账号", "name", account.AccountName, "platform", account.Platform)
	if err := db.DB.Create(&account).Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	// 自动执行一次校验
	if err := performAccountCheck(&account, c.Request.Context()); err != nil {
		slog.Error("添加账号后自动校验失败", "name", account.AccountName, "error", err)
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

	// 清理凭据中的空白字符和所有换行符（防止 HTTP Header 报错）
	account.Cookie = strings.ReplaceAll(strings.ReplaceAll(account.Cookie, "\n", ""), "\r", "")
	account.Cookie = strings.TrimSpace(account.Cookie)
	account.AuthToken = strings.ReplaceAll(strings.ReplaceAll(account.AuthToken, "\n", ""), "\r", "")
	account.AuthToken = strings.TrimSpace(account.AuthToken)

	slog.Info("更新账号", "name", account.AccountName)
	if err := db.DB.Save(&account).Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "failed to update account"})
		return
	}

	// 自动执行一次校验
	if err := performAccountCheck(&account, c.Request.Context()); err != nil {
		slog.Error("更新账号后自动校验失败", "name", account.AccountName, "error", err)
	}

	c.PureJSON(http.StatusOK, account)
}

func deleteAccount(c *gin.Context) {
	id := c.Param("id")

	// 检查是否有关联任务
	var count int64
	db.DB.Model(&db.Task{}).Where("account_id = ?", id).Count(&count)
	if count > 0 {
		slog.Error("尝试删除账号失败: 存在关联任务", "account_id", id, "task_count", count)
		c.PureJSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("该账号有关联的 %d 个任务，请先删除关联任务", count)})
		return
	}

	slog.Info("删除账号", "account_id", id)
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

	slog.Info("创建任务", "name", task.Name)
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

	slog.Info("更新任务", "name", task.Name)

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
		slog.Info("检测到关键参数变更，自动重置任务状态", "name", task.Name)
		updateData["status"] = "pending"
		updateData["message"] = ""
	}

	if err := db.DB.Model(&task).Updates(updateData).Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	// 重新加载以获取关联的 Account 信息
	db.DB.Preload("Account").First(&task, task.ID)

	// 推送更新事件
	utils.BroadcastTaskUpdate(&task)

	// 刷新调度器
	scheduler.Global.UpdateTask(task.ID, task.ScheduleMode, task.Cron)

	c.PureJSON(http.StatusOK, task)
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")
	slog.Info("删除任务", "task_id", id)

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
	slog.Info("请求运行任务", "task_id", id)

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

func runAllTasks(c *gin.Context) {
	slog.Info("请求批量运行所有任务")

	var tasks []db.Task
	// 筛选条件：1. status 不是 running; 2. message 中不包含 [Fatal]
	err := db.DB.Preload("Account").
		Where("status != ?", "running").
		Where("message NOT LIKE ? OR message IS NULL", "%[Fatal]%").
		Find(&tasks).Error

	if err != nil {
		slog.Error("获取批量运行任务列表失败", "error", err)
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tasks"})
		return
	}

	if len(tasks) == 0 {
		c.PureJSON(http.StatusOK, gin.H{"message": "没有可执行的任务", "count": 0})
		return
	}

	// 生成批次 ID（时间戳 + 随机后缀，确保唯一性）
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	batchID := fmt.Sprintf("batch_%d_%s", time.Now().UnixMilli(), hex.EncodeToString(randBytes))
	WorkerManager.RegisterBatch(batchID, len(tasks))

	count := 0
	for i := range tasks {
		task := &tasks[i]
		task.Status = "running"
		task.Stage = "Started"
		db.DB.Model(task).Updates(map[string]interface{}{
			"status": "running",
			"stage":  "Started",
		})
		utils.BroadcastTaskUpdate(task)

		WorkerManager.Submit(worker.Job{Task: task, BatchID: batchID})
		count++
	}

	utils.BroadcastStatsUpdate()
	slog.Info("批量运行任务提交完成", "batch_id", batchID, "total_triggered", count)
	c.PureJSON(http.StatusOK, gin.H{"message": fmt.Sprintf("批量执行已开启，共触发 %d 个任务", count), "count": count})
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
	// 获取全局调度开关状态 (使用 Find 避免 record not found 日志噪音)
	var enabledSetting db.Setting
	db.DB.Where("key = ?", "global_schedule_enabled").Limit(1).Find(&enabledSetting)
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
	// 使用 COALESCE 避免在无账号时 SUM 返回 NULL 导致 Scan 报错
	db.DB.Model(&db.Account{}).Where("status = 1").Select("COALESCE(SUM(capacity_used), 0)").Scan(&capacityUsed)

	var todayCompleted int64
	db.DB.Model(&db.Task{}).Where("status = ? AND DATE(last_run) = DATE('now', 'localtime')", "success").Count(&todayCompleted)

	var activeAccounts int64
	db.DB.Model(&db.Account{}).Where("status = 1").Count(&activeAccounts)

	var runningTasksList []db.Task
	// 获取：
	// 1. 正在运行的任务 (running)
	// 2. 8 秒内成功且消息不为空的任务 (success)
	// 3. 未被忽略且消息不为空的失败任务 (failed)
	db.DB.Where("status = ? OR (status = ? AND last_run >= ? AND message != '') OR (status = ? AND stage != ? AND message != '')",
		"running", "success", time.Now().Add(-8*time.Second), "failed", "Dismissed").Find(&runningTasksList)

	var recentTasks []db.Task
	// 仅显示有消息记录的任务（配合“清空日志”逻辑）
	db.DB.Where("message != ''").Order("last_run desc").Limit(15).Find(&recentTasks)

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

	db.DB.Where("key = ?", "global_schedule_enabled").Limit(1).Find(&enabledSetting)
	db.DB.Where("key = ?", "global_schedule_cron").Limit(1).Find(&cronSetting)

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

func getGlobalSettings(c *gin.Context) {
	var settings []db.Setting
	db.DB.Find(&settings)

	res := make(map[string]string)
	for _, s := range settings {
		res[s.Key] = s.Value
	}
	c.PureJSON(http.StatusOK, res)
}

func updateGlobalSettings(c *gin.Context) {
	var input map[string]string
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 提前检查 Cron 校验，避免在循环中重复或次序问题
	if cronExpr, ok := input["global_schedule_cron"]; ok {
		enabled := false
		if e, ok := input["global_schedule_enabled"]; ok {
			enabled = (e == "true")
		} else {
			var s db.Setting
			db.DB.Where("key = ?", "global_schedule_enabled").First(&s)
			enabled = (s.Value == "true")
		}

		if enabled && cronExpr != "" {
			if err := scheduler.ValidateCron(cronExpr); err != nil {
				c.PureJSON(http.StatusBadRequest, gin.H{"error": "全局 Cron 表达式格式错误: " + err.Error()})
				return
			}
		}
	}

	for k, v := range input {
		db.DB.Save(&db.Setting{Key: k, Value: v})
		// 如果更新了定时任务配置，同步更新调度器
		if k == "global_schedule_enabled" || k == "global_schedule_cron" {
			var enabledSetting db.Setting
			var cronSetting db.Setting
			db.DB.Where("key = ?", "global_schedule_enabled").First(&enabledSetting)
			db.DB.Where("key = ?", "global_schedule_cron").First(&cronSetting)
			scheduler.Global.UpdateGlobalSchedule(cronSetting.Value, enabledSetting.Value == "true")
		}
	}

	// 推送统计更新
	utils.BroadcastStatsUpdate()
	c.PureJSON(http.StatusOK, gin.H{"message": "settings updated"})
}

func testBarkNotification(c *gin.Context) {
	var input struct {
		Server  string `json:"bark_server"`
		Key     string `json:"bark_device_key"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		Level   string `json:"level"`
		Sound   string `json:"sound"`
		Icon    string `json:"icon"`
		Archive string `json:"isArchive"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	title := input.Title
	if title == "" {
		title = "测试通知"
	}
	body := input.Body
	if body == "" {
		body = "这是一条来自 UCAS 的测试推送消息。"
	}
	level := input.Level
	if level == "" {
		level = "active"
	}
	sound := input.Sound
	if sound == "" {
		sound = "birdsong.caf"
	}

	err := notify.SendBarkDirect(input.Server, input.Key, title, body, level, sound, input.Icon, input.Archive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.PureJSON(http.StatusOK, gin.H{"message": "test notification sent"})
}
