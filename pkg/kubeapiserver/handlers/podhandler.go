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
	"time"
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

func bind(pod *apiobject.Pod, node *apiobject.Node) error {
	hostIp := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == "InternalIP" {
			hostIp = addr.Address
		}
	}
	if hostIp == "" {
		return errors.New("node's InternalIP can't be found")
	}
	pod.Status.HostIp = hostIp
	return nil
}

func sendPodToNode(pod *apiobject.Pod, nodeKey string) error {
	watcher, ok := watch.WatchTable[nodeKey]
	if !ok {
		log.Warn("[keepSchedule] the nodeKey is not in the watchTable, the nodeKey: ", nodeKey)
		return errors.New("the nodeKey is not in the watchTable, the nodeKey: " + nodeKey)
	}
	jsonBytes, err := pod.MarshalJSON()
	if err != nil {
		return err
	}
	watcher.Write(jsonBytes)
	return nil
}

// keepSchedule keep check the pod's status until successfully schedule it
func keepSchedule(podKey string, nodes []apiobject.Node) {
	// use a loop, check the pod's status every 1 minute
	// if the pod's status is running, then continue
	// else break
	length := len(nodes)
	pos := 0

	for {
		time.Sleep(time.Minute * 3)

		// check current pod's status
		var pod apiobject.Pod
		err := podStorageTool.Get(context.Background(), podKey, &pod)
		if err != nil {
			log.Error("[keepSchedule] get pod failed, the key: ", podKey, "the error message: ", err.Error())
			continue
		}

		if pod.Status.Phase != apiobject.Running {
			break
		}

		// send the pod to the next node candidate
		pos = (pos + 1) % length
		nodeKey := nodes[pos].Data.Name
		sendPodToNode(&pod, nodeKey)
	}
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
	pod.Status.Phase = apiobject.Pending
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
			if ok && node.Status.Addresses != nil && len(node.Status.Addresses) != 0 {
				// TODO: the message format should be defined later
				log.Info("[CreatePod] choose the node")
				pod.Status.Phase = apiobject.Running
				pod.Status.HostIp = node.Status.Addresses[0].Address
				jsonBytes, err := pod.MarshalJSON()
				err = watcher.Write(jsonBytes)
				if err != nil {
					log.Error("[CreatePodHandler] send to the node failed")
					continue
				}
				scheduled = true
				break
			} else {
				continue
			}
		}
	}

	//// send the pod to scheduler by websocket
	//scheduler, ok := watch.WatchTable["scheduler"]
	//if ok {
	//	jsonBytes, err := pod.MarshalJSON()
	//	err = scheduler.Write(jsonBytes)
	//	if err != nil {
	//		log.Error("[CreatePodHandler] send to the node failed")
	//	}
	//
	//	// read from the scheduler util something can be read
	//	response, err := scheduler.Read()
	//	if err != nil {
	//		log.Error("[CreatePodHandler] read from the scheduler failed")
	//	}
	//	// parse the response
	//	var selectedNodes []apiobject.Node
	//	err = apiobject.UnMarshalJSONList(response, &selectedNodes)
	//	if err != nil {
	//		log.Error("[CreatePodHandler] unmarshal the response failed")
	//	}
	//	if selectedNodes == nil || len(selectedNodes) == 0 {
	//		log.Error("[CreatePodHandler] no available node")
	//	} else {
	//		scheduled = true
	//	}
	//
	//	// for convenience, api server take the duty of binding the pod to the node
	//	err = bind(pod, &selectedNodes[0])
	//	// change to running status
	//	pod.Status.Phase = apiobject.Running
	//	if err != nil {
	//		log.Error("[CreatePodHandler] bind the pod to the node failed")
	//	}
	//
	//	// write the pod to the node
	//	nodeKey := selectedNodes[0].Data.Name
	//	sendPodToNode(pod, nodeKey)
	//
	//	// keep check and resend to the next node if necessary
	//	if len(selectedNodes) > 1 {
	//		go keepSchedule(key, selectedNodes)
	//	}
	//}

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

	if pod.Status.Phase == apiobject.Running {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "the pod is running, can not delete"})
		return
	}

	// 2.2 change the pod's status
	err = changePodResourceVersion(&pod, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	pod.Status.Phase = apiobject.Terminating
	err = podStorageTool.GuaranteedUpdate(context.Background(), key, &pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2.3 delete the pod information in etcd
	err = podStorageTool.Delete(context.Background(), key)
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
