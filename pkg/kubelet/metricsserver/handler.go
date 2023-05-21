package metricsserver

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/kubelet/pod"
	"net/http"
)

type route struct {
	Path    string
	Method  string
	Handler gin.HandlerFunc
}

var handlerTable = [...]route{
	//{Path: "/:namespace/:pod", Method: "GET", Handler: nil},
	{Path: "/:namespace/:pod", Method: "GET", Handler: getPodMetricsHandler},
}

func getPodMetricsHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("pod")

	metrics := pod.GetPodMetrics(namespace, name)
	c.JSON(http.StatusOK, metrics)
}
