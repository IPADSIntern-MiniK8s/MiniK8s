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
	"fmt"
)

var jobStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

// changejobResourceVersion to change job resource version
func changeJobResourceVersion(job *apiobject.Job, c *gin.Context) error {
	method := c.Request.Method
	uri := c.Request.RequestURI
	isUpdate := strings.Contains(uri, "update")
	if method == "POST" {
		// update
		if isUpdate {
			job.Data.ResourcesVersion = apiobject.UPDATE
		} else {
			job.Data.ResourcesVersion = apiobject.CREATE
		}
	} else if method == "DELETE" {
		job.Data.ResourcesVersion = apiobject.DELETE
	} else if method != "GET" {
		// unsupported method
		return errors.New("unsupported un idempotent method")
	}
	return nil
}

// CreateJobHandler the url format is POST /api/v1/namespaces/:namespace/jobs
func CreateJobHandler(c *gin.Context) {
	// 1. parse the request to get the job object
	namespace := c.Param("namespace")
	if namespace == "" {
		// namespace is empty error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
	}

	var job *apiobject.Job
	if err := c.Bind(&job); err != nil {
		// parse body error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the job information in the storage
	key := "/registry/jobs/" + namespace + "/" + job.Data.Name
	// check whether it is a real create job request
	var prevjob apiobject.Service
	err := jobStorageTool.Get(context.Background(), key, &prevjob)
	if err == nil {
		// the job already exists
		err = updatejob(job, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, job)
		}
		return
	}
	// 2.1 change the job resource version
	if err := changeJobResourceVersion(job, c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(job)

	// 2.2 save the job information in the etcd
	key = "/registry/jobs/" + namespace + "/" + job.Data.Name
	log.Debug("[CreateJobHandler] key is ", key)

	err = jobStorageTool.Create(context.Background(), key, &job)
	if err != nil {
		log.Error("[CreateJobHandler] save job information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the job information
	c.JSON(http.StatusOK, job)
}

func updatejob(service *apiobject.Job, key string) error {
	service.Data.ResourcesVersion = apiobject.UPDATE
	err := serviceStorageTool.GuaranteedUpdate(context.Background(), key, service)
	if err != nil {
		log.Error("[UpdateServiceHandler] update service information error, ", err)
		return err
	}
	return nil
}

// GetJobHandler the url format is GET /api/v1/namespaces/:namespace/jobs/:name
func GetJobHandler(c *gin.Context) {
	// 1. parse the request to get the job object
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

	// 2. get the job information from the storage
	key := "/registry/jobs/" + namespace + "/" + name
	log.Debug("[GetJobHandler] key is ", key)

	job := &apiobject.Job{}
	err := jobStorageTool.Get(context.Background(), key, job)
	if err != nil {
		log.Error("[GetJobHandler] get job information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the job information
	c.JSON(http.StatusOK, job)
}

// GetJobsHandler the url format is GET /api/v1/namespaces/:namespace/jobs
func GetJobsHandler(c *gin.Context) {
	// 1. parse the request to get the job key
	namespace := c.Param("namespace")
	key := "/registry/jobs/" + namespace
	log.Debug("[GetJobsHandler] key is ", key)

	// 2. get the job information from the storage
	var jobList []apiobject.Job
	err := jobStorageTool.GetList(context.Background(), key, &jobList)
	if err != nil {
		log.Error("[GetJobsHandler] get job information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the job information
	c.JSON(http.StatusOK, jobList)
}

// UpdateJobHandler the url format is PUT /api/v1/namespaces/:namespace/jobs/:name/update
func UpdateJobHandler(c *gin.Context) {
	// 1. parse the request to get the job object
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

	var job *apiobject.Job
	if err := c.Bind(&job); err != nil {
		// parse body error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. save the job information in the storage
	// 2.1 change the job resource version
	if err := changeJobResourceVersion(job, c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2.2 save the job information in the etcd
	key := "/registry/jobs/" + namespace + "/" + job.Data.Name
	log.Debug("[UpdateJobHandler] key is ", key)

	err := jobStorageTool.GuaranteedUpdate(context.Background(), key, &job)
	if err != nil {
		log.Error("[UpdateJobHandler] save job information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. return the job information
	c.JSON(http.StatusOK, job)
}

// DeleteJobHandler the url format is DELETE /api/v1/namespaces/:namespace/jobs/:name
func DeleteJobHandler(c *gin.Context) {
	// 1. parse the request to get the job object
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

	// 2. delete the job information from the storage
	// actually, we just need to change the resource version
	key := "/registry/jobs/" + namespace + "/" + name
	log.Debug("[DeleteJobHandler] key is ", key)
	job := &apiobject.Job{}
	err := jobStorageTool.Get(context.Background(), key, job)
	if err != nil {
		log.Error("[DeleteJobHandler] get job information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = changeJobResourceVersion(job, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = jobStorageTool.GuaranteedUpdate(context.Background(), key, &job)
	if err != nil {
		log.Error("[DeleteJobHandler] delete job information error, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// truly delete from etcd
	err = jobStorageTool.Delete(context.Background(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	// 3. return the job information
	c.JSON(http.StatusOK, job)
}
