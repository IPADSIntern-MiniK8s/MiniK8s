package handlers

import (
	"context"
	"errors"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/storage"
	"minik8s/pkg/kubedns/nginx"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var dnsStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

func generatePath(rawPath string, host string, method string) error {
	parts := strings.Split(rawPath, ".")
	result := "/dns"
	for i := len(parts) - 1; i >= 0; i-- {
		result = result + "/" + parts[i]
	}
	log.Info("[generatePath] the new dns path is ", result)
	hoststr := apiobject.DNSEntry{
		Host: "192.168.1.13",
	}

	if method == "create" {
		err := dnsStorageTool.Create(context.Background(), result, &hoststr)
		return err
	} else {
		err := dnsStorageTool.GuaranteedUpdate(context.Background(), result, &hoststr)
		return err
	}
}

func getAllDNSRecords() []apiobject.DNSRecord {
	var dnsRecords []apiobject.DNSRecord
	err := dnsStorageTool.GetList(context.Background(), "/registry/dnsrecords", &dnsRecords)
	if err != nil {
		log.Error("[getAllDNSRecords] error getting all DNS records: ", err)
	}
	return dnsRecords
}

func getServiceAddr(serviceName string, namespace string) (string, error) {
	var service apiobject.Service
	key := "/registry/services/"
	if namespace != "" {
		key = key + namespace + "/" + serviceName
	} else {
		key = key + "default/" + serviceName
	}
	err := dnsStorageTool.Get(context.Background(), key, &service)
	if err != nil {
		log.Error("[getServiceAddr] error getting service: ", err)
		return "", err
	}
	
	if service.Data.Name == serviceName {
		if service.Status.ClusterIP != "" {
			return service.Status.ClusterIP, nil
		}
	}

	return "", errors.New("service not found")
}
	

func updateNginx() error {
	allRecord := getAllDNSRecords()
	nginx.GenerateConfig(allRecord)
	err := nginx.ReloadNginx()
	return err
}
// CreateDNSRecordHandler create a DNS record
// CreateDNSRecordHandler the url format POST /api/v1/namespaces/:namespace/dns
func CreateDNSRecordHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
		return
	}
	var dnsRecord *apiobject.DNSRecord
	if err := c.Bind(&dnsRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 1. if the address field is empty, fill it with the service address
	n := len(dnsRecord.Paths)
	for i := 0; i < n; i++ {
		if dnsRecord.Paths[i].Address == "" {
			addr, err := getServiceAddr(dnsRecord.Paths[i].Service, dnsRecord.NameSpace)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}
			dnsRecord.Paths[i].Address = addr
		}
	}
	
	// 2. save the DNSRecord and the path in the etcd
	key := "/registry/dnsrecords/" + namespace + "/" + dnsRecord.Name
	err := dnsStorageTool.Create(context.Background(), key, &dnsRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 3. save the hostname and the path in the nginx
	err = generatePath(dnsRecord.Host, "0.0.0.0", "create")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 4. update the nginx config
	err = updateNginx()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dnsRecord)
}

// UpdateDNSRecordHandler update a DNS record
// UpdateDNSRecordHandler the url format POST /api/v1/namespaces/:namespace/dns/:name/update
func UpdateDNSRecordHandler(c *gin.Context) {
	// 1. parse the DNSRecord from the request to get the name
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

	// 2. parse the DNSRecord from the request
	var dnsRecord *apiobject.DNSRecord
	if err := c.Bind(&dnsRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 3. save the DNSRecord and the path in the etcd
	key := "/registry/dnsrecords/" + namespace + "/" + name
	err := dnsStorageTool.GuaranteedUpdate(context.Background(), key, &dnsRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// save the host path in etcd
	err = generatePath(dnsRecord.Host, "0.0.0.0", "update")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 4. update the nginx config
	err = updateNginx()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	

	c.JSON(http.StatusOK, dnsRecord)
}

// DeleteDNSRecordHandler delete a DNS record
// DeleteDNSRecordHandler the url format DELETE /api/v1/namespaces/:namespace/dns/:name
func DeleteDNSRecordHandler(c *gin.Context) {
	// 1. parse the DNSRecord from the request to get the name
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

	// 2. delete the DNSRecord and the path in the etcd
	key := "/registry/dnsrecords/" + namespace + "/" + name
	err := dnsStorageTool.Delete(context.Background(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 3. update the nginx config
	err = updateNginx()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "delete success",
	})
}

// GetDNSRecordHandler get a DNS record
// GetDNSRecordHandler the url format GET /api/v1/namespaces/:namespace/dns/:name
func GetDNSRecordHandler(c *gin.Context) {
	// 1. parse the DNSRecord from the request to get the name
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

	// 2. get the DNSRecord and the path in the etcd
	key := "/registry/dnsrecords/" + namespace + "/" + name
	dnsRecord := &apiobject.DNSRecord{}
	err := dnsStorageTool.Get(context.Background(), key, dnsRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dnsRecord)
}

// GetDNSRecordsHandler list all DNS records
func GetDNSRecordsHandler(c *gin.Context) {
	var dnsRecords []apiobject.DNSRecord = getAllDNSRecords()
	c.JSON(http.StatusOK, dnsRecords)
}
