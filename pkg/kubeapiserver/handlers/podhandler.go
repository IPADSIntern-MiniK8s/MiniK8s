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

var client, _ = clientv3.New(clientv3.Config{
	Endpoints: []string{"localhost:2380"},
})

// use global variable p to store the for handle pod
var p = NewPodHandler(client)

// CreatePodHandler the url format is POST /api/v1/namespaces/:namespace/pods
// TODO: bind the pod in runtime
func CreatePodHandler(c *gin.Context) {
	// 1. parse the request get the pod from the request
	rawUrl := c.Request.URL.Path
	r, err := regexp.Compile("/api/v1/namespaces/([^/]+)/pods")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	namespace := r.FindStringSubmatch(rawUrl)[1]
	log.Debug("[CreatePodHandler] namespace: ", namespace)
	log.Debug("[CreatePodHandler] the raw url is: ", rawUrl)

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
	log.Debug("[CreatePodHandler] key: ", key)

	err = p.StorageTool.Create(context.Background(), key, jsonBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the pod to the client
	c.JSON(http.StatusOK, pod)
}

// GetPodHandler the url format is GET /api/v1/namespaces/:namespace/pods/:name
// if the request is a watch request and is a legal request, return false, nil
func GetPodHandler(c *gin.Context) {
	// 1. parse the request get the pod from the request
	rawUrl := c.Request.URL.Path
	r, err := regexp.Compile("/api/v1/namespaces/([^/]+)/pods/([^/]+)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 1.1 check whether it is a watch request
	if c.Query("watch") == "true" {
		log.Debug("[GetPodHandler] it is a watch request")
		c.Status(http.StatusSeeOther)
		return
	}
	namespace := r.FindStringSubmatch(rawUrl)[1]
	name := r.FindStringSubmatch(rawUrl)[2]
	log.Debug("[GetPodHandler] namespace: ", namespace)
	log.Debug("[GetPodHandler] name: ", name)
	log.Debug("[GetPodHandler] the raw url is: ", rawUrl)

	// 2. get the pod's information from the storage
	key := "/registry/pods/" + namespace + "/" + name
	log.Debug("[GetPodHandler] key: ", key)

	var jsonBytes []byte
	err = p.StorageTool.Get(context.Background(), key, &jsonBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the pod to the client
	var pod apiobject.Pod
	err = pod.UnmarshalJSON(jsonBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pod)
	return
}

func DeletePodHandler(c *gin.Context) {
	// 1. parse the request get the pod from the request
	rawUrl := c.Request.URL.Path
	r, err := regexp.Compile("/api/v1/namespaces/([^/]+)/pods/([^/]+)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	namespace := r.FindStringSubmatch(rawUrl)[1]
	name := r.FindStringSubmatch(rawUrl)[2]
	log.Debug("[DeletePodHandler] namespace: ", namespace)
	log.Debug("[DeletePodHandler] name: ", name)
	log.Debug("[DeletePodHandler] the raw url is: ", rawUrl)

	// 2. delete the pod's information from the storage
	key := "/registry/pods/" + namespace + "/" + name
	log.Debug("[DeletePodHandler] key: ", key)

	err = p.StorageTool.Delete(context.Background(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the pod to the client
	c.JSON(http.StatusOK, gin.H{"message": "delete pod successfully"})
}
