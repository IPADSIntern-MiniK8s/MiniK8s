package handlers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/storage"
	"minik8s/pkg/kubeapiserver/watch"
	"net/http"
	"strings"
)

var nodeStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

func changeNodeResourceVersion(node *apiobject.Node, c *gin.Context) error {
	method := c.Request.Method
	uri := c.Request.RequestURI
	isUpdate := strings.Contains(uri, "update")
	if method == "POST" {
		// update
		if isUpdate {
			node.Data.ResourcesVersion = apiobject.UPDATE
		} else {
			node.Data.ResourcesVersion = apiobject.CREATE
		}
	} else if method == "DELETE" {
		node.Data.ResourcesVersion = apiobject.DELETE
	} else if method != "GET" {
		// unsupported method
		return errors.New("unsupported un idempotent method")
	}
	return nil
}

// RegisterNodeHandler the url format is POST /api/v1/nodes
// record the node information in etcd and
// convert the connection to websocket connection
func RegisterNodeHandler(c *gin.Context) {
	// 1. get the node information from the request
	var node *apiobject.Node = &apiobject.Node{}
	log.Debug("[RegisterNodeHandler] node: get the node information from the request")
	reqBody, err := io.ReadAll(c.Request.Body)
	log.Debug("[RegisterNodeHandler] node info:", string(reqBody))
	err = node.UnMarshalJSON(reqBody)
	if err != nil {
		log.Error("get node information failed: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// change the node status to "Ready"
	node.Status = apiobject.NodeStatus{}
	if node.Status.Conditions == nil || len(node.Status.Conditions) == 0 {
		node.Status.Conditions = make([]apiobject.Condition, 1)
	}
	node.Status.Conditions[0].Status = apiobject.Ready
	// change the node's resource version
	err = changeNodeResourceVersion(node, c)
	if err != nil {
		log.Error("change node resource version failed: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. record it in etcd
	key := "/registry/nodes/" + node.Data.Name
	log.Debug("[RegisterNodeHandler] node name: ", node.Data.Name)
	log.Debug("[RegisterNodeHandler] key: ", key)

	err = nodeStorageTool.Create(context.Background(), key, &node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// put the node information in the response json
	c.JSON(http.StatusOK, node)
}

// NodeWatchHandler the url format is GET /api/v1/nodes/{name}/watch
// watch the node information
// actually, we don't use it now
func NodeWatchHandler(c *gin.Context) {
	// 1. get the node information from etcd
	name := c.Param("name")
	key := "/registry/nodes/" + name
	log.Debug("[NodeWatchHandler] key: ", key)

	node := &apiobject.Node{}
	err := nodeStorageTool.Get(context.Background(), key, node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. convert the connection to websocket connection
	// and store it in the watchTable
	watcher, err := watch.NewWatchServer(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	watch.WatchTable[key] = watcher

	watcher.Write([]byte("websocket connection is established"))
	c.JSON(http.StatusOK, node)
}

// GetNodesHandler the url format is GET /api/v1/nodes
// get a list of nodes' information from etcd
func GetNodesHandler(c *gin.Context) {
	// 1. get the node information from etcd
	key := "/registry/nodes"
	log.Debug("[GetNodesHandler] key: ", key)

	var nodeList []apiobject.Node
	err := nodeStorageTool.GetList(context.Background(), key, &nodeList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nodeList)
}

// GetNodeByNameHandler the url format is GET /api/v1/nodes/{name}
// get a node's information from etcd
func GetNodeByNameHandler(c *gin.Context) {
	// 1. get the node information from etcd
	name := c.Param("name")
	key := "/registry/nodes/" + name
	log.Debug("[GetNodeByNameHandler] key: ", key)

	node := &apiobject.Node{}
	err := nodeStorageTool.Get(context.Background(), key, node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, node)
}
