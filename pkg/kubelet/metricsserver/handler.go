package metricsserver

import (
	"github.com/gin-gonic/gin"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubelet/pod"
	kubeletUtils "minik8s/pkg/kubelet/utils"
	"minik8s/utils"
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
	info := utils.GetObject(config.POD, namespace, name)
	p := &apiobject.Pod{}
	p.UnMarshalJSON([]byte(info))

	kubeletUtils.RLock(namespace, name)
	metrics := pod.GetPodMetrics(namespace, *p)
	kubeletUtils.RUnLock(namespace, name)
	c.JSON(http.StatusOK, metrics)
}
