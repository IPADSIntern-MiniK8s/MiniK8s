package kubeproxy

import (
	"minik8s/pkg/kubeproxy/ipvs"
	"testing"
)

func TestProxy(t *testing.T) {
	ipvs.Init()
	var e proxyEndpointHandler
	var s proxyServiceHandler
	/* test add service and add endpoint */
	svcJson := "{\n    \"apiVersion\": \"v1\",\n    \"kind\": \"Service\",\n    \"metadata\": {\n        \"name\": \"service-practice\"\n    },\n    \"spec\": {\n        \"selector\": {\n            \"app\": \"deploy-1\"\n        },\n        \"type\": \"ClusterIP\",\n        \"ports\": [\n            {\n                \"name\": \"service-port1\",\n                \"protocol\": \"TCP\",\n                \"port\": 8080,\n                \"targetPort\": \"p1\"\n            }\n        ]\n    },\n    \"status\":{\n        \"ClusterIP\":\"10.10.0.2\"\n    }\n}"
	s.HandleUpdate([]byte(svcJson))
	edptJson := "{\n  \"metadata\": {\n    \"name\": \"my-service\"\n  },\n  \"spec\": {\n    \"svcIP\": \"10.10.0.2\",\n    \"svcPort\": 8080,\n    \"dstIP\": \"10.2.17.54\",\n    \"dstPort\": 12345\n  }\n}"
	edptJson2 := "{\n  \"metadata\": {\n    \"name\": \"my-service\"\n  },\n  \"spec\": {\n    \"svcIP\": \"10.10.0.2\",\n    \"svcPort\": 8080,\n    \"dstIP\": \"10.2.18.54\",\n    \"dstPort\": 12345\n  }\n}"
	e.HandleCreate([]byte(edptJson))
	e.HandleCreate([]byte(edptJson2))

	if svc, ok := ipvs.Services["10.10.0.2:8080"]; !ok {
		t.Error("Add Service Fail")
	} else {
		if _, ok := svc.Endpoints["10.2.17.54:12345"]; !ok {
			t.Error("Add Endpoint Fail")
		}
		if _, ok := svc.Endpoints["10.2.18.54:12345"]; !ok {
			t.Error("Add Endpoint Fail")
		}
	}

	/* test delete endpoint */
	e.HandleDelete([]byte(edptJson))
	svc := ipvs.Services["10.10.0.2:8080"]
	if _, ok := svc.Endpoints["10.2.17.54:12345"]; ok {
		t.Error("Add Endpoint Fail")
	}
	if _, ok := svc.Endpoints["10.2.18.54:12345"]; !ok {
		t.Error("Add Endpoint Fail")
	}

	/* test delete service */
	s.HandleDelete([]byte(svcJson))
	if _, ok := ipvs.Services["10.10.0.2:8080"]; ok {
		t.Error("Add Endpoint Fail")
	}
	e.HandleDelete([]byte(edptJson2))
}
