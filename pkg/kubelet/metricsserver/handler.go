package metricsserver

import "github.com/gin-gonic/gin"

type route struct {
	Path    string
	Method  string
	Handler gin.HandlerFunc
}

var handlerTable = [...]route{
	{Path: "/:namespace/:pod", Method: "GET", Handler: nil},
	{Path: "/:namespace/:pod", Method: "GET", Handler: nil},
}

func GetPodMetrics(c *gin.Context) {

}
