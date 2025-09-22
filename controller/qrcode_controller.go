package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

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
