package controller

import (
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver"
	"minik8s/utils"
	"testing"
)

func TestSvcController(t *testing.T) {
	utils.ApiServerIp = "localhost:8080"
	go kubeapiserver.Run()

	podJson := "{\n  \"apiVersion\": \"v1\",\n  \"kind\": \"Pod\",\n  \"metadata\": {\n    \"name\": \"example-pod1\",\n    \"labels\": {\n      \"app\": \"deploy-1\"\n    }\n  },\n  \"spec\": {\n    \"containers\": [\n      {\n        \"name\": \"example-container\",\n        \"image\": \"nginx\",\n        \"ports\": [\n          {\n            \"containerPort\": 80\n          }\n        ]\n      }\n    ]\n  },\n  \"status\": {\n    \"podIP\":\"192.168.0.1\"\n  }\n}"
	podJson2 := "{\n  \"apiVersion\": \"v1\",\n  \"kind\": \"Pod\",\n  \"metadata\": {\n    \"name\": \"example-pod2\",\n    \"labels\": {\n      \"app\": \"deploy-2\"\n    }\n  },\n  \"spec\": {\n    \"containers\": [\n      {\n        \"name\": \"example-container\",\n        \"image\": \"nginx\",\n        \"ports\": [\n          {\n            \"containerPort\": 80\n          }\n        ]\n      }\n    ]\n  },\n  \"status\": {\n    \"podIP\":\"192.168.0.2\"\n  }\n}"
	pod1 := &apiobject.Pod{}
	pod1.UnMarshalJSON([]byte(podJson))
	pod2 := &apiobject.Pod{}
	pod2.UnMarshalJSON([]byte(podJson2))
	utils.CreateObject(pod1, utils.POD, pod1.Data.Namespace)
	utils.CreateObject(pod1, utils.POD, pod2.Data.Namespace)
	/* 逻辑1：新建service，绑定已有pod，创建对应endpoint。更改service label，删除并重新创建end point。删除service， 删除已有endpoint */
}
