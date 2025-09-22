package controller

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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