package nginx

import (
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
				Address: "10.1.1.10",
				Service: "service1",
				Port:    8010,
			},
			{
				Address: "10.1.1.11",
				Service: "service2",
				Port:    8011,
			},
		},
	})
	DNSRecordList = append(DNSRecordList, apiobject.DNSRecord{
		Kind:       "DNS",
		APIVersion: "v1",
		Name:       "dns-test2",
		Host:       "node2.com",
		Paths: []apiobject.Path{
			{
				Address: "10.1.1.12",
				Service: "service3",
				Port:    8081,
			},
			{
				Address: "10.1.1.13",
				Service: "service4",
				Port:    8082,
			},
		},
	})

	GenerateConfig(DNSRecordList)
}
