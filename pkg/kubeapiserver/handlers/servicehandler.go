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

var serviceStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

// changeServiceResourceVersion to change service resource version
func changeServiceResourceVersion(service *apiobject.Service, c *gin.Context) error {
	method := c.Request.Method
	uri := c.Request.RequestURI
	isUpdate := strings.Contains(uri, "update")
	if method == "POST" {
		// update
		if isUpdate {
			service.Data.ResourcesVersion = apiobject.UPDATE
		} else {
			service.Data.ResourcesVersion = apiobject.CREATE
		}
	} else if method == "DELETE" {
		service.Data.ResourcesVersion = apiobject.DELETE
	} else if method != "GET" {
		// unsupported method
		return errors.New("unsupported un idempotent method")
	}
	return nil
}

// CreateServiceHandler the url format is POST /api/v1/namespaces/:namespace/services
func CreateServiceHandler(c *gin.Context) {
	// 1. parse the request to get the service object
	namespace := c.Param("namespace")
	if namespace == "" {
		// namespace is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
	}

	var service *apiobject.Service
	if err := c.Bind(&service); err != nil {
		// parse body error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the service information in the storage
	// 2.1 change the service resource version
	if err := changeServiceResourceVersion(service, c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2.2 save the service information in the etcd
	key := "/registry/services/" + namespace + "/" + service.Data.Name
	log.Debug("[CreateServiceHandler] key is ", key)

	err := serviceStorageTool.Create(context.Background(), key, &service)
	if err != nil {
		log.Error("[CreateServiceHandler] save service information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the service information
	c.JSON(http.StatusOK, service)
}

// GetServiceHandler the url format is GET /api/v1/namespaces/:namespace/services/:name
func GetServiceHandler(c *gin.Context) {
	// 1. parse the request to get the service object
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

	// 2. get the service information from the storage
	key := "/registry/services/" + namespace + "/" + name
	log.Debug("[GetServiceHandler] key is ", key)

	service := &apiobject.Service{}
	err := serviceStorageTool.Get(context.Background(), key, service)
	if err != nil {
		log.Error("[GetServiceHandler] get service information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the service information
	c.JSON(http.StatusOK, service)
}

// GetServicesHandler the url format is GET /api/v1/namespaces/:namespace/services
func GetServicesHandler(c *gin.Context) {
	// 1. parse the request to get the service key
	namespace := c.Param("namespace")
	key := "/registry/services/" + namespace
	log.Debug("[GetServicesHandler] key is ", key)

	// 2. get the service information from the storage
	var serviceList []apiobject.Service
	err := serviceStorageTool.GetList(context.Background(), key, &serviceList)
	if err != nil {
		log.Error("[GetServicesHandler] get service information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the service information
	c.JSON(http.StatusOK, serviceList)
}

// UpdateServiceHandler the url format is PUT /api/v1/namespaces/:namespace/services/:name/update
func UpdateServiceHandler(c *gin.Context) {
	// 1. parse the request to get the service object
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

	var service *apiobject.Service
	if err := c.Bind(&service); err != nil {
		// parse body error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the service information in the storage
	// 2.1 change the service resource version
	if err := changeServiceResourceVersion(service, c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2.2 save the service information in the etcd
	key := "/registry/services/" + namespace + "/" + service.Data.Name
	log.Debug("[UpdateServiceHandler] key is ", key)

	err := serviceStorageTool.GuaranteedUpdate(context.Background(), key, &service)
	if err != nil {
		log.Error("[UpdateServiceHandler] save service information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the service information
	c.JSON(http.StatusOK, service)
}

// DeleteServiceHandler the url format is DELETE /api/v1/namespaces/:namespace/services/:name
func DeleteServiceHandler(c *gin.Context) {
	// 1. parse the request to get the service object
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

	// 2. delete the service information from the storage
	// actually, we just need to change the resource version
	key := "/registry/services/" + namespace + "/" + name
	log.Debug("[DeleteServiceHandler] key is ", key)
	service := &apiobject.Service{}
	err := serviceStorageTool.Get(context.Background(), key, service)
	if err != nil {
		log.Error("[DeleteServiceHandler] get service information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = changeServiceResourceVersion(service, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = serviceStorageTool.GuaranteedUpdate(context.Background(), key, &service)
	if err != nil {
		log.Error("[DeleteServiceHandler] delete service information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the service information
	// TODO: maybe can't use pointer?
	c.JSON(http.StatusOK, service)
}
