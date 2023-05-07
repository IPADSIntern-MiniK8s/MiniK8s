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

var endpointStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

// changeEndpointResourceVersion change the resource version of the endpoint
func changeEndpointResourceVersion(endpoint *apiobject.Endpoint, c *gin.Context) error {
	method := c.Request.Method
	uri := c.Request.RequestURI
	isUpdate := strings.Contains(uri, "update")
	if method == "POST" {
		// update
		if isUpdate {
			endpoint.Data.ResourcesVersion = apiobject.UPDATE
		} else {
			endpoint.Data.ResourcesVersion = apiobject.CREATE
		}
	} else if method == "DELETE" {
		endpoint.Data.ResourcesVersion = apiobject.DELETE
	} else if method != "GET" {
		// unsupported method
		return errors.New("unsupported un idempotent method")
	}
	return nil
}

// CreateEndpointHandler the url format is /api/v1/namespaces/:namespace/endpoints
func CreateEndpointHandler(c *gin.Context) {
	// 1. parse the request to get the endpoint
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
		return
	}

	var endpoint *apiobject.Endpoint
	if err := c.Bind(&endpoint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 2. save the service information in the storage
	if err := changeEndpointResourceVersion(endpoint, c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// save in etcd
	key := "/registry/endpoints/" + namespace + "/" + endpoint.Data.Name
	log.Debug("[CreateEndpointHandler] key: ", key)

	err := endpointStorageTool.Create(context.Background(), key, &endpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 3. return the result
	c.JSON(http.StatusOK, endpoint)
}

// GetEndpointHandler the url format is /api/v1/namespaces/:namespace/endpoints/:name
// get the endpoint by name
func GetEndpointHandler(c *gin.Context) {
	// 1. parse the request to get the endpoint information
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
		return
	}
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
		return
	}

	// 2. get the endpoint information from the storage
	key := "/registry/endpoints/" + namespace + "/" + name
	log.Debug("[GetEndpointHandler] key: ", key)

	service := &apiobject.Endpoint{}
	err := endpointStorageTool.Get(context.Background(), key, service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 3. return the result
	c.JSON(http.StatusOK, service)
}

// GetEndpointsHandler the url format is GET /api/v1/namespaces/:namespace/endpoints
// get the endpoints by namespace
func GetEndpointsHandler(c *gin.Context) {
	// 1. parse the request to get the endpoint information
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
		return
	}

	// 2. get the endpoint information from the storage
	key := "/registry/endpoints/" + namespace
	log.Debug("[GetEndpointsHandler] key: ", key)

	var endpointList []apiobject.Endpoint
	err := endpointStorageTool.GetList(context.Background(), key, &endpointList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 3. return the result
	c.JSON(http.StatusOK, endpointList)
}

// UpdateEndpointHandler the url format is  /api/v1/namespaces/:namespace/endpoints/:name/update
// update the endpoint by name
func UpdateEndpointHandler(c *gin.Context) {
	// 1. parse the request to get the endpoint information
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
		return
	}

	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
		return
	}

	var endpoint *apiobject.Endpoint
	if err := c.Bind(&endpoint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 2. update the endpoint information in the storage
	// 2.1 change the endpoint resource version
	if err := changeEndpointResourceVersion(endpoint, c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 2.2 save in etcd
	key := "/registry/endpoints/" + namespace + "/" + name
	log.Debug("[UpdateEndpointHandler] key: ", key)

	err := endpointStorageTool.GuaranteedUpdate(context.Background(), key, &endpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 3. return the result
	c.JSON(http.StatusOK, endpoint)
}

// DeleteEndpointHandler the url format is DELETE /api/v1/namespaces/:namespace/endpoints/:name
// delete the endpoint by name
func DeleteEndpointHandler(c *gin.Context) {
	// 1. parse the request to get the endpoint information
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
		return
	}

	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
		return
	}

	// 2. delete the endpoint information in the storage
	// actually, we just need to change the resource version
	key := "/registry/endpoints/" + namespace + "/" + name
	log.Debug("[DeleteEndpointHandler] key: ", key)
	endpoint := &apiobject.Endpoint{}
	err := endpointStorageTool.Get(context.Background(), key, endpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = changeEndpointResourceVersion(endpoint, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = endpointStorageTool.GuaranteedUpdate(context.Background(), key, &endpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 3. return the result
	c.JSON(http.StatusOK, endpoint)
}
