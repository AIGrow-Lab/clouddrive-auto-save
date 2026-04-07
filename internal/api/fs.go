package api

import (
	"embed"
	"io/fs"
	"net/http"
)

// StaticFiles 将由 Dockerfile 在构建时通过编译参数或硬编码路径映射
// 这里我们预留给全局调用
var StaticFiles embed.FS

func GetStaticFS() http.FileSystem {
	// 获取 dist 目录下的内容
	sub, err := fs.Sub(StaticFiles, "dist")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}
