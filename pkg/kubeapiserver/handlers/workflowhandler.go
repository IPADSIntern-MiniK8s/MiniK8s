package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/storage"
	"minik8s/pkg/kubeapiserver/watch"
	"net/http"
)

var workFlowStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

func updateWorkflow(workflow *apiobject.WorkFlow, key string) error {
	workflow.Status = apiobject.UPDATE
	err := workFlowStorageTool.GuaranteedUpdate(context.Background(), key, workflow)
	if err != nil {
		return err
	}
	return nil
}

// UploadWorkflowHandler the url format is POST /api/v1/workflows
func UploadWorkflowHandler(c *gin.Context) {
	var workflow *apiobject.WorkFlow
	if err := c.Bind(&workflow); err != nil {
		log.Error("[UploadWorkflowHandler] parse body error: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the workflow information in the storage
	name := workflow.Name
	if name == "" {
		// workflow name is empty error
		log.Error("[UploadWorkflowHandler] workflow name is empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workflow name is empty",
		})
		return
	}

	key := "/registry/workflows/" + name

	// check whether it is a real create workflow request
	var prevWorkflow apiobject.WorkFlow
	err := workFlowStorageTool.Get(context.Background(), key, &prevWorkflow)
	if err == nil {
		// the workflow already exists
		err = updateWorkflow(workflow, key)
		if err != nil {
			log.Error("[UploadWorkflowHandler] update workflow error: ", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, workflow)
		}

		return
	}

	// 3. save the workflow information in the storage
	workflow.Status = apiobject.CREATE
	err = workFlowStorageTool.Create(context.Background(), key, &workflow)
	if err != nil {
		log.Error("[UploadWorkflowHandler] create workflow error: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workflow)
}


// DeleteWorkflowHandler the url format is DELETE /api/v1/workflows/:name
func DeleteWorkflowHandler(c *gin.Context) {
	// 1. parse the request to get the workflow name
	name := c.Param("name")
	if name == "" {
		// workflow name is empty error
		log.Error("[DeleteWorkflowHandler] workflow name is empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workflow name is empty",
		})
		return
	}

	// 2. delete the workflow information in the storage
	key := "/registry/workflows/" + name
	err := workFlowStorageTool.Delete(context.Background(), key)
	if err != nil {
		log.Error("[DeleteWorkflowHandler] delete workflow error: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "delete workflow success")
}

// GetWorkflowHandler the url format is GET /api/v1/workflows/:name
func GetWorkflowHandler(c *gin.Context) {
	// 1. parse the request to get the workflow name
	name := c.Param("name")
	if name == "" {
		// workflow name is empty error
		log.Error("[GetWorkflowHandler] workflow name is empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workflow name is empty",
		})
		return
	}

	// 2. get the workflow information in the storage
	key := "/registry/workflows/" + name
	var workflow apiobject.WorkFlow
	err := workFlowStorageTool.Get(context.Background(), key, &workflow)
	if err != nil {
		log.Error("[GetWorkflowHandler] get workflow error: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// GetWorkflowsHandler the url format is GET /api/v1/workflows
func GetWorkflowsHandler(c *gin.Context) {
	// 1. get the workflows information in the storage
	key := "/registry/workflows"
	var workflows []apiobject.WorkFlow
	err := workFlowStorageTool.GetList(context.Background(), key, &workflows)
	if err != nil {
		log.Error("[GetWorkflowsHandler] get workflows error: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workflows)
}


// TriggerWorkflowHandler the url format is POST /api/v1/workflows/:name/trigger
func TriggerWorkflowHandler(c *gin.Context) {
	// 1. parse the request to get the workflow name
	name := c.Param("name")
	if name == "" {
		// workflow name is empty error
		log.Error("[TriggerWorkflowHandler] workflow name is empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workflow name is empty",
		})
		return
	}

	// 2. get the workflow information in the storage
	key := "/registry/workflows/" + name
	var workflow apiobject.WorkFlow
	err := workFlowStorageTool.Get(context.Background(), key, &workflow)

	if err != nil {
		log.Error("[TriggerWorkflowHandler] get workflow error: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	workflowData, err := workflow.MarshalJSON()
	if err != nil {
		log.Error("[TriggerWorkflowHandler] marshal workflow error: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	log.Info("[TriggerWorkflowHandler] workflowData: ", string(workflowData))

	// 3. trigger the workflow
	handler, ok := watch.WatchTable[workflow.Name]
	if !ok {
		log.Error("[TriggerWorkflowHandler] workflow handler not found")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "workflow handler not found"})
		return
	}
	params, err := c.GetRawData()


	// the request format for trigger function is {"name": "function name", "params": "function params"}
	request := `{"name": "` + string(workflowData)  + `", "params": ` + string(params) + `}`
	log.Info("[TriggerWorkflowHandler] request: ", request)

	// 3.1 send the workflow to the handler
	err = handler.Write([]byte(request))
	if err != nil {
		log.Error("[TriggerWorkflowHandler] trigger workflow error: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3.2 wait for the result
	response, err := handler.Read()
	if err != nil {
		// read response error
		log.Error("[TriggerWorkflowHandler] read response error: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// UpdateWorkflowHandler the url format is POST /api/v1/workflows/:name/update
func UpdateWorkflowHandler(c *gin.Context) {
	// 1. parse the request to get the workflow name
	// name := c.Param("name")

	// 2. get the workflow information in the storage

}
