package metricsserver

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/kubelet/pod"
	"net/http"
	"minik8s/utils"
	"minik8s/config"
	"minik8s/pkg/apiobject"
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
	info := utils.GetObject(config.POD,namespace,name)
	p := &apiobject.Pod{}
	p.UnMarshalJSON([]byte(info))

	metrics := pod.GetPodMetrics(namespace,*p)
	c.JSON(http.StatusOK, metrics)
}
