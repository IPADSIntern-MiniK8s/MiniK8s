package handlers

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/storage"
	"net/http"
	"regexp"
)

type PodHandler struct {
	StorageTool *storage.EtcdStorage
}

func NewPodHandler(client *clientv3.Client) *PodHandler {
	return &PodHandler{
		StorageTool: storage.NewEtcdStorage(client),
	}
}

// CreatePod the url format is POST /api/v1/namespaces/:namespace/pods
// TODO: bind the pod in runtime
func (p *PodHandler) CreatePod(c *gin.Context) {
	// 1. parse the request get the pod from the request
	rawUrl := c.Request.URL.Path
	r, err := regexp.Compile("/api/v1/namespaces/([^/]+)/pods")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	namespace := r.FindStringSubmatch(rawUrl)[1]
	log.Debug("[CreatePod] namespace: ", namespace)
	log.Debug("[CreatePod] the raw url is: ", rawUrl)

	var pod *apiobject.Pod
	if err = c.Bind(&pod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the pod's information in the storage
	jsonBytes, err := pod.MarshalJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	key := "/registry/pods/" + namespace + "/" + pod.Data.Name
	log.Debug("[CreatePod] key: ", key)

	err = p.StorageTool.Create(context.Background(), key, jsonBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the pod to the client
	c.JSON(http.StatusOK, pod)
}
