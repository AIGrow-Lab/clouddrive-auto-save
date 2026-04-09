package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	_ "github.com/zcq/clouddrive-auto-save/internal/core/cloud139"
	_ "github.com/zcq/clouddrive-auto-save/internal/core/quark"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
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
		// ... 现有 API 路由
		api.GET("/accounts", listAccounts)
		api.POST("/accounts", createAccount)
		api.PUT("/accounts/:id", updateAccount)
		api.DELETE("/accounts/:id", deleteAccount)
		api.POST("/accounts/:id/check", checkAccount)

		api.GET("/tasks", listTasks)
		api.POST("/tasks", createTask)
		api.PUT("/tasks/:id", updateTask)
		api.DELETE("/tasks/:id", deleteTask)
		api.POST("/tasks/:id/run", runTask)
	}

	// 静态资源处理
	r.NoRoute(func(c *gin.Context) {
		// 如果不是 /api 开头的请求，则尝试返回静态文件
		http.FileServer(GetStaticFS()).ServeHTTP(c.Writer, c.Request)
	})

	return r
}

func checkAccount(c *gin.Context) {
	id := c.Param("id")
	var account db.Account
	if err := db.DB.First(&account, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	driver := core.GetDriver(&account)
	if driver == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "driver not found"})
		return
	}

	// 执行实际的登录校验
	ctx := c.Request.Context()
	updatedAccount, err := driver.GetInfo(ctx)
	if err != nil {
		account.Status = 0
		db.DB.Save(&account)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 更新数据库中的信息
	db.DB.Model(&account).Updates(map[string]interface{}{
		"nickname":       updatedAccount.Nickname,
		"status":         1,
		"capacity_used":  updatedAccount.CapacityUsed,
		"capacity_total": updatedAccount.CapacityTotal,
		"vip_name":       updatedAccount.VipName,
		"last_check":     time.Now(),
	})

	c.JSON(http.StatusOK, account)
}

func listAccounts(c *gin.Context) {
	var accounts []db.Account
	db.DB.Find(&accounts)
	c.JSON(http.StatusOK, accounts)
}

func createAccount(c *gin.Context) {
	var account db.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&account)
	c.JSON(http.StatusOK, account)
}

func updateAccount(c *gin.Context) {
	id := c.Param("id")
	var account db.Account
	if err := db.DB.First(&account, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Save(&account)
	c.JSON(http.StatusOK, account)
}

func deleteAccount(c *gin.Context) {
	id := c.Param("id")
	db.DB.Delete(&db.Account{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func listTasks(c *gin.Context) {
	var tasks []db.Task
	db.DB.Preload("Account").Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}

func createTask(c *gin.Context) {
	var task db.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&task)
	c.JSON(http.StatusOK, task)
}

func updateTask(c *gin.Context) {
	id := c.Param("id")
	var task db.Task
	if err := db.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Save(&task)
	c.JSON(http.StatusOK, task)
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")
	db.DB.Delete(&db.Task{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}

func runTask(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var task db.Task
	if err := db.DB.Preload("Account").First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	if task.Status == "running" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task is already running"})
		return
	}

	// 提交到全局 Worker 管理器
	WorkerManager.Submit(worker.Job{Task: &task})

	c.JSON(http.StatusOK, gin.H{"message": "task submitted to worker pool"})
}
