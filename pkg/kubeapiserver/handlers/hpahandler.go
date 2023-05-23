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

var hpaStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

// changeHpaResourceVersion to change hpa resource version
func changeHpaResourceVersion(hpa *apiobject.HorizontalPodAutoscaler, c *gin.Context) error {
	method := c.Request.Method
	uri := c.Request.RequestURI
	isUpdate := strings.Contains(uri, "update")
	if method == "POST" {
		// update
		if isUpdate {
			hpa.Data.ResourcesVersion = apiobject.UPDATE
		} else {
			hpa.Data.ResourcesVersion = apiobject.CREATE
		}
	} else if method == "DELETE" {
		hpa.Data.ResourcesVersion = apiobject.DELETE
	} else if method != "GET" {
		// unsupported method
		return errors.New("unsupported un idempotent method")
	}
	return nil
}

// CreateHpaHandler the url format is POST /api/v1/namespaces/:namespace/hpas
func CreateHpaHandler(c *gin.Context) {
	// 1. parse the request to get the hpa object
	namespace := c.Param("namespace")
	if namespace == "" {
		// namespace is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
	}

	var hpa *apiobject.HorizontalPodAutoscaler
	if err := c.Bind(&hpa); err != nil {
		// parse body error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the hpa information in the storage
	key := "/registry/hpas/" + namespace + "/" + hpa.Data.Name
	// check whether it is a real create hpa request
	var prevHpa apiobject.HorizontalPodAutoscaler
	err := hpaStorageTool.Get(context.Background(), key, &prevHpa)
	if err == nil {
		// already exist
		c.JSON(http.StatusConflict, gin.H{
			"error": "hpa already exist",
		})
		return
	}

	// change the resource version
	if err := changeHpaResourceVersion(hpa, c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// save the hpa
	if err := hpaStorageTool.Create(context.Background(), key, hpa); err != nil {
		log.Error("[CreateHpaHandler] create hpa error: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the hpa object
	c.JSON(http.StatusOK, hpa)
}

// GetHpaHandler the url format is GET /api/v1/namespaces/:namespace/hpas/:name
func GetHpaHandler(c *gin.Context) {
	// 1. parse the request to get the hpa object
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

	// 2. get the hpa information from the storage
	key := "/registry/hpas/" + namespace + "/" + name
	hpa := &apiobject.HorizontalPodAutoscaler{}
	err := hpaStorageTool.Get(context.Background(), key, hpa)
	if err != nil {
		log.Error("[GetHpaHandler] get hpa error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the hpa object
	c.JSON(http.StatusOK, hpa)
}

// GetHpasHandler the url format is GET /api/v1/namespaces/:namespace/hpas
func GetHpasHandler(c *gin.Context) {
	// 1. parse the request to get the hpa object
	namespace := c.Param("namespace")
	if namespace == "" {
		// namespace is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
	}

	// 2. get the hpa information from the storage
	key := "/registry/hpas/" + namespace
	var hpas []apiobject.HorizontalPodAutoscaler
	err := hpaStorageTool.GetList(context.Background(), key, &hpas)
	if err != nil {
		log.Error("[GetHpasHandler] get hpa error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the hpa object
	c.JSON(http.StatusOK, hpas)
}

func updateHpa(hpa *apiobject.HorizontalPodAutoscaler, key string) error {
	hpa.Data.ResourcesVersion = apiobject.UPDATE
	if err := hpaStorageTool.GuaranteedUpdate(context.Background(), key, hpa); err != nil {
		log.Error("update hpa error: ", err)
		return err
	}
	return nil
}

// UpdateHpaHandler the url format is POST /api/v1/namespaces/:namespace/hpas/:name
func UpdateHpaHandler(c *gin.Context) {
	namespaces := c.Param("namespace")
	if namespaces == "" {
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

	var hpa *apiobject.HorizontalPodAutoscaler
	if err := c.Bind(&hpa); err != nil {
		// parse body error
		log.Error("[UpdateHpaHandler] parse body error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the hpa information in the storage
	key := "/registry/hpas/" + namespaces + "/" + name

	if err := updateHpa(hpa, key); err != nil {
		log.Error("[UpdateHpaHandler] update hpa error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the hpa object
	c.JSON(http.StatusOK, hpa)
}

// DeleteHpaHandler the url format is DELETE /api/v1/namespaces/:namespace/hpas/:name
func DeleteHpaHandler(c *gin.Context) {
	// 1. parse the request to get the hpa object
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

	key := "/registry/hpas/" + namespace + "/" + name
	log.Debug("[DeleteHpaHandler] key: ", key)
	hpa := &apiobject.HorizontalPodAutoscaler{}
	err := hpaStorageTool.Get(context.Background(), key, hpa)
	if err != nil {
		log.Error("[DeleteHpaHandler] get hpa error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = changeHpaResourceVersion(hpa, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. update in the storage
	err = hpaStorageTool.GuaranteedUpdate(context.Background(), key, &hpa)
	if err != nil {
		log.Error("[DeleteHpaHandler] delete hpa error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// truly delete from the storage
	err = hpaStorageTool.Delete(context.Background(), key)
	if err != nil {
		log.Error("[DeleteHpaHandler] delete hpa error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the hpa object
	c.JSON(http.StatusOK, hpa)
}

// GetAllHpaHandler the url format is GET /api/v1/hpas
func GetAllHpaHandler(c *gin.Context) {
	key := "/registry/hpas"
	var hpas []apiobject.HorizontalPodAutoscaler
	err := hpaStorageTool.GetList(context.Background(), key, &hpas)
	if err != nil {
		log.Error("[GetAllHpaHandler] get hpa error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the hpa object
	c.JSON(http.StatusOK, hpas)
}