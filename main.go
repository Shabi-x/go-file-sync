package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/zserge/lorca"
)

func main() {
	args := []string{
		"--remote-allow-origins=*",  // 关键参数：允许所有来源的远程连接
		"--remote-debugging-port=9222", // 固定调试端口
		"--no-sandbox",              // 禁用沙盒，解决权限问题
		"--disable-setuid-sandbox",  // 禁用setuid沙盒
		"--disable-dev-shm-usage",   // 解决/dev/shm使用问题
		"--disable-accelerated-2d-canvas", // 禁用2D画布加速
		"--no-first-run",            // 跳过首次运行
		"--disable-default-apps",    // 禁用默认应用
	}
	
	ui, _ := lorca.New("data:text/html,<html><head><title>go-file-sync</title></head><body><h1>Hello, Lorca!</h1></body></html>", "", 800, 600, args...)
	defer ui.Close()

	// 等待信号
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigc:
	case <-ui.Done():
	}
}

