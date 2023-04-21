package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/api/v1/namespaces/default/pods", func(c *gin.Context) {
		data, _ := c.GetRawData()
		fmt.Printf("receive: %s", string(data))
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	fmt.Print("run")
	r.Run("127.0.0.1:8080") // 监听并在 0.0.0.0:8080 上启动服务
}
