package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/core/renamer"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"net/http"
	"regexp"
	"time"
)

type PreviewRequest struct {
	AccountID   uint       `json:"account_id" binding:"required"`
	ShareURL    string     `json:"share_url" binding:"required"`
	ExtractCode string     `json:"extract_code"`
	StartDate   *time.Time `json:"start_date"`
	Pattern     string     `json:"pattern"`
	Replacement string     `json:"replacement"`
	Name        string     `json:"name"` // 任务名称，用于变量替换
}

type PreviewResult struct {
	OriginalName string `json:"original_name"`
	NewName      string `json:"new_name"`
	Matched      bool   `json:"matched"`
	IsFiltered   bool   `json:"is_filtered"`
}

func previewTask(c *gin.Context) {
	var req PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var account db.Account
	if err := db.DB.First(&account, req.AccountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	driver := core.GetDriver(&account)
	if driver == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "driver not found"})
		return
	}

	ctx := c.Request.Context()
	files, err := driver.ParseShare(ctx, req.ShareURL, req.ExtractCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 实例化重命名引擎
	renameProc := renamer.NewProcessor()

	var results []PreviewResult
	for _, f := range files {
		res := PreviewResult{
			OriginalName: f.Name,
			NewName:      f.Name,
		}

		// 日期过滤逻辑
		if req.StartDate != nil && !f.UpdateTime.IsZero() && f.UpdateTime.Before(*req.StartDate) {
			res.IsFiltered = true
			results = append(results, res)
			continue
		}

		// 重命名逻辑
		if req.Pattern != "" || req.Replacement != "" {
			opts := renamer.RenameOptions{
				TaskName:    req.Name,
				Pattern:     req.Pattern,
				Replacement: req.Replacement,
				FileName:    f.Name,
			}
			newName, err := renameProc.Process(opts)
			if err == nil {
				res.NewName = newName
				// 简单的匹配状态判断：如果新名不等于旧名，或者正则能匹配上
				if req.Pattern != "" {
					re, _ := regexp.Compile(req.Pattern)
					res.Matched = re != nil && re.MatchString(f.Name)
				} else {
					res.Matched = newName != f.Name
				}
			}
		}

		results = append(results, res)
	}

	c.JSON(http.StatusOK, results)
}
