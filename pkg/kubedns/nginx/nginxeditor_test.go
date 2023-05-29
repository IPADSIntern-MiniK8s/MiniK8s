package nginx

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/apiobject"
	"testing"
)

func TestNginxEditor(t *testing.T) {
	DNSRecordList := make([]apiobject.DNSRecord, 0)
	DNSRecordList = append(DNSRecordList, apiobject.DNSRecord{
		Kind:       "DNS",
		APIVersion: "v1",
		Name:       "dns-test1",
		Host:       "node1.com",
		Paths: []apiobject.Path{
			{
				PathName: "path1",
				Address: "127.1.1.10",
				Service: "service1",
				Port:    8010,
			},
			{
				PathName: "path2",
				Address: "127.1.1.11",
				Service: "service2",
				Port:    8011,
			},
		},
	})
	jsonData, err := json.MarshalIndent(DNSRecordList[0], "", "    ")
	if err != nil {
		fmt.Println("Error marshalling data:", err)
		return
	}
	t.Log("json data: ", string(jsonData))
	// fmt.Println(string(jsonData))
	DNSRecordList = append(DNSRecordList, apiobject.DNSRecord{
		Kind:       "DNS",
		APIVersion: "v1",
		Name:       "dns-test2",
		Host:       "node2.com",
		Paths: []apiobject.Path{
			{
				Address: "127.1.1.12",
				Service: "service3",
				Port:    8081,
			},
			{
				Address: "127.1.1.13",
				Service: "service4",
				Port:    8082,
			},
		},
	})

	GenerateConfig(DNSRecordList)
}
