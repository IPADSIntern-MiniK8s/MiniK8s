package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/storage"
	"minik8s/pkg/kubeapiserver/watch"
	"net/http"
	"regexp"
)

var podStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

func changePodStatus(pod *apiobject.Pod, status string) error {
	pod.Status.Phase = status
	key := "registry/pods/" + pod.Data.Namespace + "/" + pod.Data.Name
	err := podStorageTool.GuaranteedUpdate(context.Background(), key, pod)
	if err != nil {
		return err
	}
	return nil
}

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
	// set the pod status
	pod.Status.Phase = "Pending"
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	key := "/registry/pods/" + namespace + "/" + pod.Data.Name
	log.Debug("[CreatePodHandler] key: ", key)

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

	var pod apiobject.Pod
	err = podStorageTool.Get(context.Background(), key, &pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the pod to the client
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

	err = podStorageTool.Delete(context.Background(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the pod to the client
	c.JSON(http.StatusOK, gin.H{"message": "delete pod successfully"})
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

	// 2. update the node information in etcd
	key := "/registry/pods/" + namespace + "/" + name
	log.Debug("[UpdatePodStatusHandler] key: ", key)
	err := podStorageTool.GuaranteedUpdate(context.Background(), key, &pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the result to the client
	c.JSON(http.StatusOK, pod)
}
