package main

import (
	"embed"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"github.com/zserge/lorca"
)

//go:embed frontend/dist/*
var FS embed.FS

func main() {
	// 启动Gin服务器，启动一个协程
	go func() {
		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()
		r.GET("/", func(c *gin.Context) {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, "<html><head><title>go-file-sync</title></head><body><h1>Hello, Gin!</h1><p>This is served by Gin on port 8080</p></body></html>")
		})
		r.POST("/api/v1/texts", TextController)
		r.GET("/uploads/:path", UploadsController)
		r.GET("/api/v1/addresses", AddressController)
		r.GET("/api/v1/qrcodes", QRCodesController)
		r.POST("/api/v1/files", FilesController)
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

/**
 * @api {post} /api/v1/texts 文本上传
 * @apiName UploadText
 * @apiGroup Text
 *
 * @apiParam {String} raw 文本内容
 *
 * @apiSuccess {String} filename 文件名
 */
func TextController(c *gin.Context) {
	var json struct {
		Raw string `json:"raw"` // 前端传递的文本内容
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
	} else {
		exe, err := os.Executable() // 获取当前可执行文件的路径,并绑定到uploads目录下
		if err != nil {
			log.Fatal("get executable path failed", err)
		}
		dir := filepath.Dir(exe) // 获取当前可执行文件exe的目录
		if err != nil {
			log.Fatal("get executable dir failed", err)
		}
		fileName := uuid.New().String()
		uploads := filepath.Join(dir, "uploads") //exe 所在目录拼接uploads目录
		err = os.MkdirAll(uploads, os.ModePerm)  // 创建了一个我们完全能控制的uploads目录, 权限是0777
		if err != nil {
			log.Fatal("create uploads dir failed", err)
		}
		fullPath := path.Join(uploads, fileName+".txt")
		err = os.WriteFile(fullPath, []byte(json.Raw), os.ModePerm) // 将json.Raw写入返回的fullPath文件中
		if err != nil {
			log.Fatal("write file failed", err)
		}
		c.JSON(http.StatusOK, gin.H{"filename": fileName + ".txt"})
	}
}

/**
 * @api {get} /api/v1/addresses 获取IP地址并返回给前端
 * @apiName GetAddresses
 * @apiGroup Address
 *
 * @apiSuccess {String[]} addresses IP地址列表
 */
func AddressController(c *gin.Context) {
	addrs, _ := net.InterfaceAddrs() // 获取所有ip地址
	var result []string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				result = append(result, ipnet.IP.String())
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"addresses": result})
}

func getUploadDir() (uploads string) {
	exe, err := os.Executable() // 获取当前可执行文件的路径,并绑定到uploads目录下
	if err != nil {
		log.Fatal("get executable path failed", err)
	}
	dir := filepath.Dir(exe) // 获取当前可执行文件exe的目录
	uploads = filepath.Join(dir, "uploads")
	return
}

/**
 * @api {get} /uploads/:path 文件下载
 * @apiName DownloadFile
 * @apiGroup File
 *
 * @apiParam {String} path 文件路径（相对路径）
 *
 * @apiSuccess {File} file 下载的文件
 * @apiError {NotFound} 404 文件未找到
 */
func UploadsController(c *gin.Context) {
	// 从URL参数中获取文件路径
	if path := c.Param("path"); path != "" {
		// 构建完整的文件路径（上传目录 + 文件名）
		target := filepath.Join(getUploadDir(), path)
		// 设置HTTP响应头
		c.Header("Content-Description", "File Transfer")              // 内容描述：文件传输
		c.Header("Content-Transfer-Encoding", "binary")               // 传输编码：二进制
		c.Header("Content-Type", "application/octet-stream")          // 内容类型：二进制流，支持所有文件类型
		c.Header("Content-Disposition", "attachment; filename="+path) // 内容处置：作为附件下载，并指定文件名

		// 将文件发送给前端
		c.File(target)
	} else {
		c.Status(http.StatusNotFound)
	}
}

/**
 * @api {get} /api/v1/qrcodes 获取二维码
 * @apiName GetQRCode
 * @apiGroup QRCode
 *
 * @apiParam {String} content 前端传递的二维码url链接
 *
 * @apiSuccess {File} file 下载的文件
 */
func QRCodesController(c *gin.Context) {
	if content := c.Query("content"); content != "" {
		png, err := qrcode.Encode(content, qrcode.Medium, 256)
		if err != nil {
			log.Fatal("encode qrcode failed", err)
		}
		c.Data(http.StatusOK, "image/png", png)
	} else {
		c.Status(http.StatusBadRequest)
	}
}

func FilesController(c *gin.Context) {
	file, err := c.FormFile("raw")
	if err != nil {
		log.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Dir(exe)
	if err != nil {
		log.Fatal(err)
	}
	filename := uuid.New().String()
	uploads := filepath.Join(dir, "uploads")
	err = os.MkdirAll(uploads, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	fullpath := path.Join("uploads", filename+filepath.Ext(file.Filename))
	fileErr := c.SaveUploadedFile(file, filepath.Join(dir, fullpath))
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	c.JSON(http.StatusOK, gin.H{"url": "/" + fullpath})
}
