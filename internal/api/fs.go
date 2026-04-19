//go:build !embed

package api

import (
	"net/http"
)

// GetStaticFS 本地开发模式：直接从本地目录读取资源
func GetStaticFS() http.FileSystem {
	return http.Dir("web/dist")
}
