package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

func getAccountFolders(c *gin.Context) {
	id := c.Param("id")
	parentID := c.Query("parent_id")
	parentPath := c.Query("parent_path")

	slog.Info("正在获取账号目录树", "account_id", id, "parent_id", parentID, "parent_path", parentPath)

	var account db.Account
	if err := db.DB.First(&account, id).Error; err != nil {
		slog.Error("获取目录失败: 账号未找到", "account_id", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	driver := core.GetDriver(&account)
	if driver == nil {
		slog.Error("获取目录失败: 驱动加载失败", "platform", account.Platform)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Driver not found"})
		return
	}

	folders, err := driver.ListFiles(c.Request.Context(), parentID)
	if err != nil {
		slog.Error("获取目录异常", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 仅保留文件夹并转换为前端 Tree 组件需要的格式
	var result []map[string]interface{}
	for _, f := range folders {
		if f.IsFolder {
			// 基于 parent_path 拼接出完整的绝对路径，确保前端 save_path 显示的是人类可读的路径
			absPath := f.Name
			if parentPath != "" && parentPath != "/" && parentPath != "0" && parentPath != "root" {
				absPath = parentPath + "/" + f.Name
			} else {
				absPath = "/" + f.Name
			}

			result = append(result, map[string]interface{}{
				"id":     f.ID,
				"name":   f.Name,
				"label":  f.Name, // 增加 label 字段供前端 Tree 组件使用
				"path":   absPath,
				"isLeaf": false, // 统一设为 false 以支持继续展开
			})
		}
	}

	slog.Info("获取目录完成", "account_id", id, "folder_count", len(result))
	c.JSON(http.StatusOK, result)
}

func createAccountFolder(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name       string `json:"name"`
		ParentPath string `json:"parent_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Info("正在创建目录", "account_id", id, "name", req.Name, "parent_path", req.ParentPath)

	var account db.Account
	if err := db.DB.First(&account, id).Error; err != nil {
		slog.Error("创建文件夹失败: 账号未找到", "account_id", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	driver := core.GetDriver(&account)
	if driver == nil {
		slog.Error("创建文件夹失败: 驱动加载失败", "platform", account.Platform)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Driver not found"})
		return
	}

	info, err := driver.CreateFolder(c.Request.Context(), req.ParentPath, req.Name)
	if err != nil {
		slog.Error("创建文件夹异常", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	childPath := req.Name
	if req.ParentPath != "" && req.ParentPath != "/" && req.ParentPath != "0" && req.ParentPath != "root" {
		childPath = req.ParentPath + "/" + req.Name
	} else {
		childPath = "/" + req.Name
	}

	slog.Info("创建文件夹完成", "path", childPath)
	c.JSON(http.StatusOK, gin.H{
		"id":     info.ID,
		"name":   info.Name,
		"label":  info.Name,
		"path":   childPath,
		"isLeaf": false,
	})
}
