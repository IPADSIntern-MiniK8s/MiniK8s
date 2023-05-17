package handlers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/storage"
	"net/http"
	"strings"
)

var replicaStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

// changeReplicaResourceVersion to change replica resource version
func changeReplicaResourceVersion(replica *apiobject.ReplicationController, c *gin.Context) error {
	method := c.Request.Method
	uri := c.Request.RequestURI
	isUpdate := strings.Contains(uri, "update")
	if method == "POST" {
		// update
		if isUpdate {
			replica.Data.ResourcesVersion = apiobject.UPDATE
		} else {
			replica.Data.ResourcesVersion = apiobject.CREATE
		}
	} else if method == "DELETE" {
		replica.Data.ResourcesVersion = apiobject.DELETE
	} else if method != "GET" {
		// unsupported method
		return errors.New("unsupported un idempotent method")
	}
	return nil
}

// CreateReplicaHandler the url format is POST /api/v1/namespaces/:namespace/replicas
func CreateReplicaHandler(c *gin.Context) {
	// 1. parse the request to get the replica object
	namespace := c.Param("namespace")
	if namespace == "" {
		// namespace is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
	}

	var replica *apiobject.ReplicationController
	if err := c.Bind(&replica); err != nil {
		// parse body error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the replica information in the storage
	// 2.1 change the replica resource version
	if err := changeReplicaResourceVersion(replica, c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2.2 save the replica information in the etcd
	key := "/registry/replicas/" + namespace + "/" + replica.Data.Name
	log.Debug("[CreateReplicaHandler] key is ", key)

	err := replicaStorageTool.Create(context.Background(), key, &replica)
	if err != nil {
		log.Error("[CreateReplicaHandler] save replica information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the replica information
	c.JSON(http.StatusOK, replica)
}

// GetReplicaHandler the url format is GET /api/v1/namespaces/:namespace/replicas/:name
func GetReplicaHandler(c *gin.Context) {
	// 1. parse the request to get the replica object
	namespace := c.Param("namespace")
	if namespace == "" {
		// namespace is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
	}

	name := c.Param("name")
	if name == "" {
		// name is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
	}

	// 2. get the replica information from the storage
	key := "/registry/replicas/" + namespace + "/" + name
	log.Debug("[GetReplicaHandler] key is ", key)

	replica := &apiobject.ReplicationController{}
	err := replicaStorageTool.Get(context.Background(), key, replica)
	if err != nil {
		log.Error("[GetReplicaHandler] get replica information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the replica information
	c.JSON(http.StatusOK, replica)
}

// GetReplicasHandler the url format is GET /api/v1/namespaces/:namespace/replicas
func GetReplicasHandler(c *gin.Context) {
	// 1. parse the request to get the replica key
	namespace := c.Param("namespace")
	key := "/registry/replicas/" + namespace
	log.Debug("[GetReplicasHandler] key is ", key)

	// 2. get the replica information from the storage
	var replicaList []apiobject.ReplicationController
	err := replicaStorageTool.GetList(context.Background(), key, &replicaList)
	if err != nil {
		log.Error("[GetReplicasHandler] get replica information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the replica information
	c.JSON(http.StatusOK, replicaList)
}

// UpdateReplicaHandler the url format is PUT /api/v1/namespaces/:namespace/replicas/:name/update
func UpdateReplicaHandler(c *gin.Context) {
	// 1. parse the request to get the replica object
	namespace := c.Param("namespace")
	if namespace == "" {
		// namespace is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
	}

	name := c.Param("name")
	if name == "" {
		// name is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
	}

	var replica *apiobject.ReplicationController
	if err := c.Bind(&replica); err != nil {
		// parse body error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the replica information in the storage
	// 2.1 change the replica resource version
	if err := changeReplicaResourceVersion(replica, c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2.2 save the replica information in the etcd
	key := "/registry/replicas/" + namespace + "/" + replica.Data.Name
	log.Debug("[UpdateReplicaHandler] key is ", key)

	err := replicaStorageTool.GuaranteedUpdate(context.Background(), key, &replica)
	if err != nil {
		log.Error("[UpdateReplicaHandler] save replica information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the replica information
	c.JSON(http.StatusOK, replica)
}

// DeleteReplicaHandler the url format is DELETE /api/v1/namespaces/:namespace/replicas/:name
func DeleteReplicaHandler(c *gin.Context) {
	// 1. parse the request to get the replica object
	namespace := c.Param("namespace")
	if namespace == "" {
		// namespace is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
	}

	name := c.Param("name")
	if name == "" {
		// name is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
	}

	// 2. delete the replica information from the storage
	// actually, we just need to change the resource version
	key := "/registry/replicas/" + namespace + "/" + name
	log.Debug("[DeleteReplicaHandler] key is ", key)
	replica := &apiobject.ReplicationController{}
	err := replicaStorageTool.Get(context.Background(), key, replica)
	if err != nil {
		log.Error("[DeleteReplicaHandler] get replica information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = changeReplicaResourceVersion(replica, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = replicaStorageTool.GuaranteedUpdate(context.Background(), key, &replica)
	if err != nil {
		log.Error("[DeleteReplicaHandler] delete replica information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// truly delete from etcd
	err = replicaStorageTool.Delete(context.Background(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	// 3. return the replica information
	c.JSON(http.StatusOK, replica)
}
