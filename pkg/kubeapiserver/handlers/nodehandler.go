package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/storage"
	"minik8s/pkg/kubeapiserver/watch"
	"net/http"
)

var nodeStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

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

	// 2. record it in etcd
	log.Debug("[RegisterNodeHandler] node: record it in etcd")
	jsonBytes, err := node.MarshalJSON()
	log.Debug("[RegisterNodeHandler] the node data:", node)
	if err != nil {
		log.Error("record node information failed: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	key := "/registry/nodes/" + node.Data.Name
	log.Debug("[RegisterNodeHandler] node name: ", node.Data.Name)
	log.Debug("[RegisterNodeHandler] key: ", key)

	err = nodeStorageTool.Create(context.Background(), key, string(jsonBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, node)
}

// NodeWatchHandler the url format is GET /api/v1/nodes/{name}/watch
// watch the node information
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

	nodeList := &apiobject.NodeList{}

	err := nodeStorageTool.GetList(context.Background(), key, &nodeList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
