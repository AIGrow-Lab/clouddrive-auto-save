package api

import (
	"log/slog"
	"net/http"
	"regexp"
	"sort"

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
		ParentID    string `json:"parent_id"`
		SavePath    string `json:"save_path"`
		Pattern     string `json:"pattern"`
		Replacement string `json:"replacement"`
		Name        string `json:"name"`
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
	files, err := driver.ParseShare(c.Request.Context(), req.ShareURL, req.ExtractCode, req.ParentID)
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
			"id":            f.ID,
			"original_name": f.Name,
			"is_folder":     f.IsFolder,
			"size":          f.Size,
			"updated":       f.UpdatedAt,
		}

		// 默认状态
		res["is_filtered"] = false

		// 如果设置了替换规则，计算预览效果
		if req.Replacement != "" {
			taskName := req.Name
			if taskName == "" {
				taskName = "Task" // 如果未输入任务名，默认使用 Task 占位预览
			}

			// 判定是否匹配
			matched := true
			if req.Pattern != "" {
				re, err := regexp.Compile(req.Pattern)
				if err == nil {
					matched = re.MatchString(f.Name)
				}
			}
			res["matched"] = matched

			if matched {
				newName, err := processor.Process(renamer.RenameOptions{
					TaskName:    taskName,
					FileName:    f.Name,
					Pattern:     req.Pattern,
					Replacement: req.Replacement,
				})
				if err == nil {
					res["new_name"] = newName
					res["is_renamed"] = newName != f.Name
				} else {
					res["rename_error"] = err.Error()
					res["new_name"] = f.Name
					res["matched"] = false
				}
			} else {
				res["new_name"] = f.Name
				res["is_renamed"] = false
			}
		} else {
			res["new_name"] = f.Name
			res["is_renamed"] = false
			res["matched"] = true
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
		ParentID    string `json:"parent_id"`
		SavePath    string `json:"save_path"`
		Pattern     string `json:"pattern"`
		Replacement string `json:"replacement"`
		Name        string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Info("正在解析分享链接", "url", req.ShareURL, "account_id", req.AccountID, "parent_id", req.ParentID, "save_path", req.SavePath)

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

	// 1. 获取分享内容
	files, err := driver.ParseShare(c.Request.Context(), req.ShareURL, req.ExtractCode, req.ParentID)
	if err != nil {
		slog.Error("解析失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. 排序：按更新时间从新到旧
	sort.Slice(files, func(i, j int) bool {
		return files[i].UpdateTime.After(files[j].UpdateTime)
	})

	// 3. 获取目标目录列表，准备去重 Map
	existingNames := make(map[string]bool)
	targetID, err := driver.PrepareTargetPath(c.Request.Context(), req.SavePath)
	if err == nil {
		existingFiles, err := driver.ListFiles(c.Request.Context(), targetID)
		if err == nil {
			for _, f := range existingFiles {
				existingNames[f.Name] = true
			}
		}
	}

	// 4. 重命名预览与去重比对
	processor := renamer.NewProcessor()
	for i := range files {
		f := &files[i]

		// 默认预览名
		f.NewName = f.Name

		// 如果有规则，计算预览名
		if req.Replacement != "" {
			taskName := req.Name
			if taskName == "" {
				taskName = "Task"
			}

			// 判定是否匹配
			matched := true
			if req.Pattern != "" {
				re, err := regexp.Compile(req.Pattern)
				if err == nil {
					matched = re.MatchString(f.Name)
				}
			}

			if matched {
				newName, err := processor.Process(renamer.RenameOptions{
					TaskName:    taskName,
					FileName:    f.Name,
					Pattern:     req.Pattern,
					Replacement: req.Replacement,
				})
				if err == nil {
					f.NewName = newName
				}
			}
		}

		// 同名预检：拿 NewName 去比对
		if existingNames[f.NewName] {
			f.IsExisted = true
		}
	}

	slog.Info("解析完成", "item_count", len(files))
	c.JSON(http.StatusOK, files)
}
