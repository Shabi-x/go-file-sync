package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/Shabix/go-file-sync/server/controller"
	"github.com/gin-gonic/gin"
)

//go:embed frontend/dist/*
var FS embed.FS

// 服务器端口配置
const port = "27149"

// StartServer 启动Gin服务器
func StartServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, "<html><head><title>go-file-sync</title></head><body><h1>Hello, Gin!</h1><p>This is served by Gin on port 8080</p></body></html>")
	})
	r.POST("/api/v1/texts", controller.TextController)
	r.GET("/uploads/:path", controller.UploadsController)
	r.GET("/api/v1/addresses", controller.AddressController)
	r.GET("/api/v1/qrcodes", controller.QRCodesController)
	r.POST("/api/v1/files", controller.FilesController)
	staticFiles, _ := fs.Sub(FS, "frontend/dist")
	r.StaticFS("/static", http.FS(staticFiles)) // 在路由中添加静态文件前端服务
	// 所有未匹配的路由都返回index.html
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/static") {
			reader, err := staticFiles.Open("index.html")
			if err != nil {
				log.Fatal(err)
			}
			defer reader.Close() // 确保文件在读取后关闭
			stat, err := reader.Stat()
			if err != nil {
				log.Fatal(err)
			}
			c.DataFromReader(http.StatusOK, stat.Size(), "text/html; charset=utf-8", reader, nil)
		} else {
			c.Status(http.StatusNotFound)
		}
	})
	r.Run(":" + port)
}
