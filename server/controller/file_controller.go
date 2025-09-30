package controller

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

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