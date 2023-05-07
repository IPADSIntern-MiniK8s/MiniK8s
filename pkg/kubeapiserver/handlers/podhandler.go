package handlers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/storage"
	"minik8s/pkg/kubeapiserver/watch"
	"net/http"
	"strings"
)

var podStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

// change the pod's resourceVersion to different value
func changePodResourceVersion(pod *apiobject.Pod, c *gin.Context) error {
	method := c.Request.Method
	uri := c.Request.RequestURI
	isUpdate := strings.Contains(uri, "update")
	if method == "POST" {
		// update
		if isUpdate {
			pod.Data.ResourcesVersion = apiobject.UPDATE
		} else {
			pod.Data.ResourcesVersion = apiobject.CREATE
		}
	} else if method == "DELETE" {
		pod.Data.ResourcesVersion = apiobject.DELETE
	} else if method != "GET" {
		// unsupported method
		return errors.New("unsupported un idempotent method")
	}
	return nil
}

// CreatePodHandler the url format is POST /api/v1/namespaces/:namespace/pods
// TODO: bind the pod in runtime
func CreatePodHandler(c *gin.Context) {
	// 1. parse the request get the pod from the request
	namespace := c.Param("namespace")
	log.Debug("[CreatePodHandler] namespace: ", namespace)

	var pod *apiobject.Pod
	if err := c.Bind(&pod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the pod's information in the storage
	// 2.1 set the pod status
	pod.Status.Phase = "Pending"
	key := "/registry/pods/" + namespace + "/" + pod.Data.Name
	log.Debug("[CreatePodHandler] key: ", key)

	// 2.2 change the pod's resourceVersion
	err := changePodResourceVersion(pod, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = podStorageTool.Create(context.Background(), key, &pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. check the node information and get the node's ip
	nodeKey := "/registry/nodes/"
	var nodeList []apiobject.Node
	err = podStorageTool.GetList(context.Background(), nodeKey, &nodeList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: this task should be executed by scheduler
	scheduled := false

	for _, node := range nodeList {
		if node.Status.Conditions[0].Status == "Ready" {
			nodeKey = node.Data.Name
			println("the watchTable is: ", watch.WatchTable, "the length is: ", len(watch.WatchTable))
			log.Debug("[CreatePodHandler] the nodeKey is: ", nodeKey)
			// print the watchTable keys
			for k, _ := range watch.WatchTable {
				println("the key is: ", k)
			}
			watcher, ok := watch.WatchTable[nodeKey]
			if ok {
				// TODO: the message format should be defined later
				pod.Status.Phase = "Running"
				jsonBytes, err := pod.MarshalJSON()
				err = watcher.Write(jsonBytes)
				if err != nil {
					log.Debug("[CreatePodHandler] send to the node failed")
					continue
				}
				scheduled = true
				break
			} else {
				continue
			}
		}
	}

	if !scheduled {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no available node"})
		return
	}

	// 4. return the pod to the client
	c.JSON(http.StatusOK, pod)
}

// GetPodHandler the url format is GET /api/v1/namespaces/:namespace/pods/:name
// if the request is a watch request and is a legal request, return false, nil
func GetPodHandler(c *gin.Context) {
	// 1. parse the request get the pod from the request
	namespace := c.Param("namespace")
	name := c.Param("name")
	log.Debug("[GetPodHandler] namespace: ", namespace)
	log.Debug("[GetPodHandler] name: ", name)

	// 2. get the pod's information from the storage
	key := "/registry/pods/" + namespace + "/" + name
	log.Debug("[GetPodHandler] key: ", key)

	var pod apiobject.Pod
	err := podStorageTool.Get(context.Background(), key, &pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the pod to the client
	c.JSON(http.StatusOK, pod)
}

// GetPodsHandler the url format is GET /api/v1/namespaces/:namespace/pods
func GetPodsHandler(c *gin.Context) {
	// 1. parse the request get the pod from the request
	namespace := c.Param("namespace")
	log.Debug("[GetPodsHandler] namespace: ", namespace)

	// 2. query the pods' information from the storage
	key := "/registry/pods/" + namespace
	log.Debug("[GetPodsHandler] key: ", key)
	var podList []apiobject.Pod
	err := podStorageTool.GetList(context.Background(), key, &podList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the pods to the client
	c.JSON(http.StatusOK, podList)
}

// DeletePodHandler the url format is DELETE /api/v1/namespaces/:namespace/pods/:name
func DeletePodHandler(c *gin.Context) {
	// 1. parse the request get the pod from the request
	namespace := c.Param("namespace")
	name := c.Param("name")
	log.Debug("[DeletePodHandler] namespace: ", namespace)
	log.Debug("[DeletePodHandler] name: ", name)

	// 2. delete the pod's information from the storage
	// use lazy delete, just change the pod's status
	key := "/registry/pods/" + namespace + "/" + name
	log.Debug("[DeletePodHandler] key: ", key)
	var pod apiobject.Pod
	err := podStorageTool.Get(context.Background(), key, &pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if pod.Status.Phase == "Running" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "the pod is running, can not delete"})
		return
	}

	// 2.2 change the pod's status
	err = changePodResourceVersion(&pod, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	pod.Status.Phase = "Terminating"
	err = podStorageTool.GuaranteedUpdate(context.Background(), key, &pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the pod to the client
	c.JSON(http.StatusOK, pod)
}

// UpdatePodStatusHandler the url format is POST /api/v1/nodes/{name}/update
// update the node's status in etcd
func UpdatePodStatusHandler(c *gin.Context) {
	// 1. parse the request get the pod from the request
	namespace := c.Param("namespace")
	name := c.Param("name")
	log.Debug("[UpdatePodStatusHandler] namespace: ", namespace)
	log.Debug("[UpdatePodStatusHandler] name: ", name)

	var pod *apiobject.Pod
	if err := c.Bind(&pod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. update the pod information in etcd
	key := "/registry/pods/" + namespace + "/" + name
	log.Debug("[UpdatePodStatusHandler] key: ", key)

	// 2.2 change the pod's status
	err := changePodResourceVersion(pod, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = podStorageTool.GuaranteedUpdate(context.Background(), key, &pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the result to the client
	c.JSON(http.StatusOK, pod)
}
