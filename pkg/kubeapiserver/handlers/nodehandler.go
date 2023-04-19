package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/apimachinery"
	"minik8s/pkg/kubeapiserver/storage"
	"net/http"
)

var nodeStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

// RegisterNodeHandler the url format is POST /api/v1/nodes
// record the node information in etcd and
// convert the connection to websocket connection
func RegisterNodeHandler(c *gin.Context) {
	// 1. get the node information from the request
	var node *apiobject.Node
	if err := c.Bind(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. record it in etcd
	jsonBytes, err := node.MarshalJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	key := "/registry/nodes/" + node.Data.Name
	log.Debug("[RegisterNodeHandler] key: ", key)

	err = nodeStorageTool.Create(context.Background(), key, string(jsonBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. convert the connection to websocket connection
	// and store it in the watchTable
	watcher, err := apimachinery.NewWatchServer(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	apimachinery.WatchTable[key] = watcher

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
