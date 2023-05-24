package handlers

import (
	"context"
	"encoding/json"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/storage"
	"minik8s/pkg/kubeapiserver/watch"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)
var functionStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()


func updateFunction(function *apiobject.Function, key string) error {
	function.Status = apiobject.UPDATE
	err := functionStorageTool.GuaranteedUpdate(context.Background(), key, function)
	if err != nil {
		return err
	}
	return nil
}

// getWatchFeedback get the watch feedback
func getWatchFeedback(c *gin.Context, prefix string, name string) {
	handler, ok := watch.WatchTable[name]
	if !ok {
		// watch table error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no according function handler",
		})
		return
	}

	for {
		response, err := handler.Read()
		log.Info("[getWatchFeedback] watch response: ", string(response))
		if err != nil {
			// read response error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		if strings.Contains(string(response), prefix) {
			log.Info("[getWatchFeedback] response: ", string(response))
			c.JSON(http.StatusOK, string(response))
			break
		}
	}
	
}

// UploadFunctionHandler the url format is POST /api/v1/functions
func UploadFunctionHandler(c *gin.Context) {
	// 1. parse the request to get the function object
	var function *apiobject.Function
	if err := c.Bind(&function); err != nil {
		// parse body error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the function information in the storage
	name := function.Name
	if name == "" {
		// function name is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "function name is empty",
		})
		return
	}

	

	key := "/registry/functions/" + name
	log.Info("[UploadFunctionHandler] key: ", key)
	// check whether it is a real create function request
	var prevFunction apiobject.Function
	err := functionStorageTool.Get(context.Background(), key, &prevFunction)
	if err == nil {
		// the function already exists
		// update the function
		err = updateFunction(function, key)
		if err != nil {
			// update function error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})	
		} else {
			getWatchFeedback(c, "update:", "function")
		}
		return
	}

	// 2.2 change the status
	function.Status = apiobject.CREATE

	// 3. save the function information in the storage
	err = functionStorageTool.Create(context.Background(), key, function)
	if err != nil {
		// save function error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 4. create the image for the function through watch
	getWatchFeedback(c, "create:", "function")
}
	

// GetFunctionHandler the url format is GET /api/v1/functions/:name
func GetFunctionHandler(c *gin.Context) {
	// 1. parse the request to get the function object
	name := c.Param("name")
	if name == "" {
		// function name is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "function name is empty",
		})
		return
	}

	key := "/registry/functions/" + name
	var function apiobject.Function
	err := functionStorageTool.Get(context.Background(), key, &function)
	if err != nil {
		// get function error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, function)
}

// DeleteFunctionHandler the url format is DELETE /api/v1/functions/:name
func DeleteFunctionHandler(c *gin.Context) {
	// 1. parse the request to get the function object
	name := c.Param("name")
	if name == "" {
		// function name is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "function name is empty",
		})
		return
	}

	key := "/registry/functions/" + name
	var function apiobject.Function
	err := functionStorageTool.Get(context.Background(), key, &function)
	if err != nil {
		// get function error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 2. change the status
	function.Status = apiobject.DELETE
	err = functionStorageTool.GuaranteedUpdate(context.Background(), key, function)
	if err != nil {
		// update function error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})	
		return
	}

	// 2. delete the function information in the storage
	err = functionStorageTool.Delete(context.Background(), key)
	if err != nil {
		// delete function error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	getWatchFeedback(c, "delete:", "function")
}

// UpdateFunctionHandler the url format is POST /api/v1/functions/:name/update
func UpdateFunctionHandler(c *gin.Context) {
	// 1. parse the request to get the function object
	name := c.Param("name")
	if name == "" {
		// function name is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "function name is empty",
		})
		return
	}

	key := "/registry/functions/" + name
	var function apiobject.Function
	err := functionStorageTool.Get(context.Background(), key, &function)
	if err != nil {
		// get function error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 2. change the status
	err = updateFunction(&function, key)
	if err != nil {
		// update function error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	getWatchFeedback(c, "update:", "function")
}


// TriggerFunctionHandler the url format is POST /api/v1/functions/:name/trigger
// the body is the parameters of the function
func TriggerFunctionHandler(c *gin.Context) {
	// 1. parse the request to get the function object
	name := c.Param("name")

	// 2. check whether the function exists
	key := "/registry/functions/" + name
	var function apiobject.Function
	err := functionStorageTool.Get(context.Background(), key, &function)
	if err != nil {
		// get function error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 3. trigger the function
	handler, ok := watch.WatchTable["function"]
	if !ok {
		// watch table error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no according handler",
		})
		return
	}

	params, err := c.GetRawData()

	// the request format for trigger function is {"name": "function name", "params": "function params"}
	request := struct {
		Name   string          `json:"name"`
		Params json.RawMessage `json:"params"`
	}{
		Name:   name,
		Params: params,
	}

	reqStr, err := json.Marshal(request)
	if err != nil {
		// get raw data error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Info("[TriggerFunctionHandler] reqStr: ", string(reqStr))

	// 3.1 send the function trigger request to the handler
	err = handler.Write(reqStr)
	if err != nil {
		// send request error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// wait for the result
	for {
		response, err := handler.Read()
		if err != nil {
			// read response error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		log.Info("[TriggerFunctionHandler] watch response: ", string(response))
		if strings.Contains(string(response), "execute:") {
			log.Info("[TriggerFunctionHandler] response: ", string(response))
			c.JSON(http.StatusOK, string(response))
			break
		}
	}
	
}


// GetFunctionsHandler the url format is GET /api/v1/functions
func GetFunctionsHandler(c *gin.Context) {
	// 1. parse the request to get the function object
	var functions []apiobject.Function
	err := functionStorageTool.GetList(context.Background(), "/registry/functions/", &functions)
	if err != nil {
		// get function error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, functions)
}