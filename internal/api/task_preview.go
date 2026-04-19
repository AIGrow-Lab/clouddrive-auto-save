package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/core/renamer"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

func previewTask(c *gin.Context) {
	var req struct {
		AccountID   uint   `json:"account_id"`
		ShareURL    string `json:"share_url"`
		ExtractCode string `json:"extract_code"`
		SavePath    string `json:"save_path"`
		Pattern     string `json:"pattern"`
		Replacement string `json:"replacement"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var account db.Account
	if err := db.DB.First(&account, req.AccountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	driver := core.GetDriver(&account)
	if driver == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Driver not found"})
		return
	}

	// 1. 获取分享内容
	files, err := driver.ParseShare(c.Request.Context(), req.ShareURL, req.ExtractCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Parse share failed: " + err.Error()})
		return
	}

	// 2. 准备重命名器
	processor := renamer.NewProcessor()
	var results []map[string]interface{}

	for _, f := range files {
		// 构造预览结果
		res := map[string]interface{}{
			"id":        f.ID,
			"old_name":  f.Name,
			"is_folder": f.IsFolder,
			"size":      f.Size,
			"updated":   f.UpdatedAt,
		}

		// 如果设置了重命名规则，计算新名字
		if req.Replacement != "" {
			newName, err := processor.Process(renamer.RenameOptions{
				TaskName:    "Preview", // 预览时临时任务名
				FileName:    f.Name,
				Pattern:     req.Pattern,
				Replacement: req.Replacement,
			})
			if err == nil {
				res["new_name"] = newName
				res["is_renamed"] = newName != f.Name
			} else {
				res["rename_error"] = err.Error()
			}
		} else {
			res["new_name"] = f.Name
			res["is_renamed"] = false
		}

		results = append(results, res)
	}

	slog.Info("预览计算完成", "file_count", len(results))
	c.JSON(http.StatusOK, results)
}

func parseShareLinkInfo(c *gin.Context) {
	var req struct {
		AccountID   uint   `json:"account_id"`
		ShareURL    string `json:"share_url"`
		ExtractCode string `json:"extract_code"`
		SavePath    string `json:"save_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Info("正在解析分享链接", "url", req.ShareURL, "account_id", req.AccountID, "save_path", req.SavePath)

	var account db.Account
	if err := db.DB.First(&account, req.AccountID).Error; err != nil {
		slog.Error("解析失败: 账号未找到", "account_id", req.AccountID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	driver := core.GetDriver(&account)
	if driver == nil {
		slog.Error("解析失败: 驱动加载失败", "platform", account.Platform)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Driver not found"})
		return
	}

	files, err := driver.ParseShare(c.Request.Context(), req.ShareURL, req.ExtractCode)
	if err != nil {
		slog.Error("解析失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 针对 139，如果 SavePath 包含通配符或需要自动匹配，我们在此可以做预处理
	// 目前仅简单返回列表

	// 过滤并排序（可选，前端通常会处理）
	var result []core.FileInfo = files

	// 针对 Quark 的特殊处理：如果用户提供了 SavePath 且包含 ID，我们可以在此进行映射
	// ... (现有逻辑暂无)

	slog.Info("解析完成", "item_count", len(files))
	c.JSON(http.StatusOK, result)
}
