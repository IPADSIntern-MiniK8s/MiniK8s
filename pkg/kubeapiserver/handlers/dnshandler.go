package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/storage"
	"net/http"
	"strings"
)

var dnsStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()

// generate the path for the DNSRecord Path
func generatePath(rawPath string, host string, method string) error {
	parts := strings.Split(rawPath, ".")
	result := "/dns"
	for i := len(parts) - 1; i >= 0; i-- {
		result = result + "/" + parts[i]
	}
	log.Info("[generatePath] the new dns path is ", result)
	hoststr := `{"host":"` + host + `"}`

	if method == "create" {
		err := dnsStorageTool.Create(context.Background(), result, hoststr)
		return err
	} else {
		err := dnsStorageTool.GuaranteedUpdate(context.Background(), result, hoststr)
		return err
	}
}

// CreateDNSRecordHandler create a DNS record
// CreateDNSRecordHandler the url format POST /api/v1/dns
func CreateDNSRecordHandler(c *gin.Context) {
	var dnsRecord *apiobject.DNSRecord
	if err := c.Bind(&dnsRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 2. save the DNSRecord and the path in the etcd
	key := "/registry/dnsrecords/" + dnsRecord.Name
	err := dnsStorageTool.Create(context.Background(), key, &dnsRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, path := range dnsRecord.Paths {
		err = generatePath(path.Address, dnsRecord.Host, "create")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, dnsRecord)
}

// UpdateDNSRecordHandler update a DNS record
// UpdateDNSRecordHandler the url format POST /api/v1/dns/:name/update
func UpdateDNSRecordHandler(c *gin.Context) {
	// 1. parse the DNSRecord from the request to get the name
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
	key := "/registry/dnsrecords/" + name
	err := dnsStorageTool.GuaranteedUpdate(context.Background(), key, &dnsRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, path := range dnsRecord.Paths {
		err = generatePath(path.Address, dnsRecord.Host, "update")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, dnsRecord)
}

// DeleteDNSRecordHandler delete a DNS record
// DeleteDNSRecordHandler the url format DELETE /api/v1/dns/:name
func DeleteDNSRecordHandler(c *gin.Context) {
	// 1. parse the DNSRecord from the request to get the name
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
		return
	}

	// 2. delete the DNSRecord and the path in the etcd
	key := "/registry/dnsrecords/" + name
	err := dnsStorageTool.Delete(context.Background(), key)
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
// GetDNSRecordHandler the url format GET /api/v1/dns/:name
func GetDNSRecordHandler(c *gin.Context) {
	// 1. parse the DNSRecord from the request to get the name
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
		return
	}

	// 2. get the DNSRecord and the path in the etcd
	key := "/registry/dnsrecords/" + name
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
	key := "/registry/dnsrecords"
	var dnsRecords []apiobject.DNSRecord
	err := dnsStorageTool.GetList(context.Background(), key, &dnsRecords)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dnsRecords)
}
