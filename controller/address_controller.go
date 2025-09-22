package controller

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
