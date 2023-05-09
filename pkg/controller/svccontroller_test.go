package controller

import (
	"minik8s/pkg/apiobject"
	"minik8s/utils"
	"testing"
)

func TestSvcController(t *testing.T) {
	utils.ApiServerIp = "localhost:8080"

	podJson1 := "{\n  \"apiVersion\": \"v1\",\n  \"kind\": \"Pod\",\n  \"metadata\": {\n    \"name\": \"example-pod1\",\n    \"labels\": {\n      \"app\": \"deploy-1\"\n    }\n  },\n  \"spec\": {\n    \"containers\": [\n      {\n        \"name\": \"example-container\",\n        \"image\": \"nginx\",\n        \"ports\": [\n          {\n            \"containerPort\": 12345,\n            \"name\": \"p1\"\n          }\n        ]\n      }\n    ]\n  },\n  \"status\":{\n      \"podIP\":\"10.2.17.54\"\n  }}"
	podJson2 := "{\n  \"apiVersion\": \"v1\",\n  \"kind\": \"Pod\",\n  \"metadata\": {\n    \"name\": \"example-pod2\",\n    \"labels\": {\n      \"app\": \"deploy-2\"\n    }\n  },\n  \"spec\": {\n    \"containers\": [\n      {\n        \"name\": \"example-container\",\n        \"image\": \"nginx\",\n        \"ports\": [\n          {\n            \"containerPort\": 12345,\n            \"name\": \"p1\"\n          }\n        ]\n      }\n    ]\n  },\n  \"status\":{\n      \"podIP\":\"10.2.17.55\"\n  }}"
	pod1 := &apiobject.Pod{}
	pod1.UnMarshalJSON([]byte(podJson1))
	pod2 := &apiobject.Pod{}
	pod2.UnMarshalJSON([]byte(podJson2))

	utils.CreateObject(pod1, utils.POD, pod1.Data.Namespace)
	utils.CreateObject(pod2, utils.POD, pod2.Data.Namespace)

	/* 逻辑1：新建service，绑定已有pod，创建对应endpoint。更改service label，删除并重新创建end point。删除service， 删除已有endpoint */
	var s svcServiceHandler

	println(" ========== Check Service Create ========= ")
	svcJson := "{\n    \"apiVersion\": \"v1\",\n    \"kind\": \"Service\",\n    \"metadata\": {\n        \"name\": \"service-practice\",\n        \"resourcesVersion\": \"UPDATE\"\n    },\n    \"spec\": {\n        \"selector\": {\n            \"app\": \"deploy-1\"\n        },\n        \"type\": \"ClusterIP\",\n        \"ports\": [\n            {\n                \"name\": \"service-port1\",\n                \"protocol\": \"TCP\",\n                \"port\": 8080,\n                \"targetPort\": \"p1\"\n            }\n        ],\n        \"clusterIP\": \"10.10.0.2\"\n    }\n}"
	svc := &apiobject.Service{}
	svc.UnMarshalJSON([]byte(svcJson))
	utils.CreateObject(svc, utils.SERVICE, svc.Data.Namespace)
	s.HandleCreate([]byte(svcJson))
	checkSvc := utils.GetObject(utils.SERVICE, "default", "service-practice")
	println(checkSvc, "\n")
	checkEdpt := utils.GetObject(utils.ENDPOINT, "default", "")
	println(checkEdpt, "\n")

	println("========== Check Service Update ========= ")
	svc.UnMarshalJSON([]byte(checkSvc))
	svc.Spec.Selector["app"] = "deploy-2"
	utils.UpdateObject(svc, utils.SERVICE, svc.Data.Namespace, svc.Data.Name)
	res, _ := svc.MarshalJSON()
	s.HandleUpdate(res)
	checkEdpt = utils.GetObject(utils.ENDPOINT, "default", "")
	println(checkEdpt, "\n")
	//s.HandleUpdate([]byte(svcJson))

	println("========== Check Service Delete ========= ")
	utils.DeleteObject(utils.SERVICE, svc.Data.Namespace, svc.Data.Name)
	s.HandleDelete(res)
	checkEdpt = utils.GetObject(utils.ENDPOINT, "default", "")
	println(checkEdpt, "\n")

	if _, ok := svcList["10.10.0.0"]; ok {
		t.Error("Service Status Fail")
	}
	utils.DeleteObject(utils.POD, "default", pod1.Data.Name)
	utils.DeleteObject(utils.POD, "default", pod2.Data.Name)

	/* 逻辑2：创建pod，对应service增加endpoint。更改pod，对应serivice改变endpoint。 */
	var p svcPodHandler
	utils.CreateObject(svc, utils.SERVICE, svc.Data.Namespace)
	s.HandleCreate([]byte(res))
	podJson1 = "{\n  \"apiVersion\": \"v1\",\n  \"kind\": \"Pod\",\n  \"metadata\": {\n    \"name\": \"example-pod1\",\n    \"labels\": {\n      \"app\": \"deploy-2\"\n    }\n  },\n  \"spec\": {\n    \"containers\": [\n      {\n        \"name\": \"example-container\",\n        \"image\": \"nginx\",\n        \"ports\": [\n          {\n            \"containerPort\": 12345,\n            \"name\": \"p1\"\n          }\n        ]\n      }\n    ]\n  },\n  \"status\":{\n      \"podIP\":\"10.2.17.54\"\n  }}"
	pod1 = &apiobject.Pod{}
	pod1.UnMarshalJSON([]byte(podJson1))
	utils.UpdateObject(pod1, utils.POD, pod1.Data.Namespace, pod1.Data.Name)
	utils.UpdateObject(pod2, utils.POD, pod2.Data.Namespace, pod2.Data.Name)

	println("========== Check Pod Create ========= ")
	p.HandleUpdate([]byte(podJson1))
	p.HandleUpdate([]byte(podJson2))
	checkEdpt = utils.GetObject(utils.ENDPOINT, "default", "")
	println(checkEdpt, "\n")
	if list, ok := svcToEndpoints["10.10.0.1"]; !ok {
		t.Error("[Check Pod Create] svcToEndpoints[\"10.10.0.1:8080\"] doesn't exist")
	} else {
		if len(*list) != 2 {
			t.Error("[Check Pod Create] fail")
		}
	}

	println("========== Check Pod Update ========= ")
	pod2.Data.Labels["app"] = "deploy-1"
	res, _ = pod2.MarshalJSON()
	p.HandleUpdate(res)
	checkEdpt = utils.GetObject(utils.ENDPOINT, "default", "")
	println(checkEdpt, "\n")
	if list, ok := svcToEndpoints["10.10.0.1"]; !ok {
		t.Error("[Check Pod Update] svcToEndpoints[\"10.10.0.1:8080\"] doesn't exist")
	} else {
		if len(*list) != 1 {
			t.Error("[Check Pod Update] fail")
		}
	}

	println("========== Check Pod Delete ========= ")
	p.HandleDelete([]byte(podJson2))
	p.HandleDelete([]byte(podJson1))
	checkEdpt = utils.GetObject(utils.ENDPOINT, "default", "")
	println(checkEdpt, "\n")
	if list, ok := svcToEndpoints["10.10.0.1"]; !ok {
		t.Error("[Check Pod Delete] svcToEndpoints[\"10.10.0.1:8080\"] doesn't exist")
	} else {
		if len(*list) != 0 {
			t.Error("[Check Pod Delete] fail")
		}
	}

}
