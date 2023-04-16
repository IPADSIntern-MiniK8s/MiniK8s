package handlers

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/storage"
	"net/http"
)

type PodHandler struct {
	StorageTool *storage.EtcdStorage
}

func NewPodHandler(client *clientv3.Client) *PodHandler {
	return &PodHandler{
		StorageTool: storage.NewEtcdStorage(client),
	}
}

// CreatePod the url format is POST /api/v1/namespaces/{namespace}/pods
func CreatePod(c *gin.Context) {
	// 1. parse the request get the pod from the request
	var pod *apiobject.Pod
	if err := c.Bind(&pod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 2. create the pod in the storage
	namespace := c.Param("namespace")
	log.Info("namespace: ", namespace)
	// 3. return the pod to the client
}
