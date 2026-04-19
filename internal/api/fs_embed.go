//go:build embed

package api

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var staticEmbed embed.FS

// GetStaticFS 获取内嵌的前端资源
func GetStaticFS() http.FileSystem {
	sub, err := fs.Sub(staticEmbed, "dist")
	if err != nil {
		// 理论上如果 embed 成功，dist 一定存在
		panic(err)
	}
	return http.FS(sub)
}
