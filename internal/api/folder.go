package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"net/http"
)

// FolderItem 为前端 TreeSelect 提供的结构
type FolderItem struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Label  string `json:"label"`
	IsLeaf bool   `json:"isLeaf"`
}

func getAccountFolders(c *gin.Context) {
	fmt.Printf("DEBUG: getAccountFolders parentID=%s parentPath=%s\n", c.Query("parent_id"), c.Query("parent_path"))
	id := c.Param("id")
	parentID := c.DefaultQuery("parent_id", "")
	parentPath := c.DefaultQuery("parent_path", "/")

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

	ctx := c.Request.Context()
	files, err := driver.ListFiles(ctx, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var folders []FolderItem
	for _, f := range files {
		if f.IsFolder {
			childPath := parentPath
			if childPath == "/" {
				childPath = "/" + f.Name
			} else {
				childPath = childPath + "/" + f.Name
			}
			folders = append(folders, FolderItem{
				ID:     f.ID, // 明确使用 ID
				Path:   childPath,
				Label:  f.Name,
				IsLeaf: false,
			})
		}
	}

	c.JSON(http.StatusOK, folders)
}

func createAccountFolder(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ParentID   string `json:"parent_id"`
		ParentPath string `json:"parent_path"`
		Name       string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	ctx := c.Request.Context()
	newFolder, err := driver.CreateFolder(ctx, req.ParentID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	childPath := req.ParentPath
	if childPath == "/" || childPath == "" {
		childPath = "/" + newFolder.Name
	} else {
		childPath = childPath + "/" + newFolder.Name
	}

	c.JSON(http.StatusOK, FolderItem{
		ID:     newFolder.ID,
		Path:   childPath,
		Label:  newFolder.Name,
		IsLeaf: false,
	})
}
