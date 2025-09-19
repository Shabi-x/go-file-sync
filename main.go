package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zserge/lorca"
)

//go:embed frontend/dist/*
var FS embed.FS

func main() {
	// 启动Gin服务器，启动一个协程
	go func() {
		r := gin.Default()
		r.GET("/", func(c *gin.Context) {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, "<html><head><title>go-file-sync</title></head><body><h1>Hello, Gin!</h1><p>This is served by Gin on port 8080</p></body></html>")
		})
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
				defer reader.Close()
				stat, err := reader.Stat()
				if err != nil {
					log.Fatal(err)
				}
				c.DataFromReader(http.StatusOK, stat.Size(), "text/html; charset=utf-8", reader, nil)
			} else {
				c.Status(http.StatusNotFound)
			}
		})
		r.Run(":8080")
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	args := []string{
		"--remote-allow-origins=*",        // 关键参数：允许所有来源的远程连接
		"--remote-debugging-port=9222",    // 固定调试端口
		"--no-sandbox",                    // 禁用沙盒，解决权限问题
		"--disable-setuid-sandbox",        // 禁用setuid沙盒
		"--disable-dev-shm-usage",         // 解决/dev/shm使用问题
		"--disable-accelerated-2d-canvas", // 禁用2D画布加速
		"--no-first-run",                  // 跳过首次运行
		"--disable-default-apps",          // 禁用默认应用
	}

	ui, _ := lorca.New("http://localhost:8080/static/", "", 800, 600, args...)
	// 等待信号, 收到signal时interrupt
	sigc := make(chan os.Signal, 1)
	// 监听system interrupt信号和terminate信号
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	select {
	// ”<-语法“是用于提取channel中的值，没有值时会阻塞当前线程，直到有预期信号之一抵达才会执行ui的关闭
	case <-sigc:
	case <-ui.Done():
	}
	ui.Close()
}
