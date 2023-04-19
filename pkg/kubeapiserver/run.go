package kubeapiserver

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/kubeapiserver/apimachinery"
)

func main() {
	myAPI := apimachinery.NewAPI()
	myAPI.RegisterHandler("GET", "/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, World!"})
	})

	err := myAPI.Run(":8080")
	if err != nil {
		panic(err)
	}
}
