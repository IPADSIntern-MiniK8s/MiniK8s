package controller

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"minik8s/pkg/apiobject"
	"minik8s/utils"
	"strconv"
)

/* 主要工作：
1. 监听service资源的创建。一旦service资源创建，为其分配唯一的cluster ip。
2. 遍历pod列表，找到符合selector条件的pod，记录。创建endpoint。
3. 监听pod创建。增加endpoint。
4. 监听pod删除。删除endpoint。
5. 监听pod更新。如果标签更改，删除/增加endpoint。
6. 监听service资源的删除。删除对应endpoint。
*/

var IPMap = [1 << 8]bool{false}
var IPStart = "10.10.0."

var svcToEndpoints = map[string]*[]*apiobject.Endpoint{}
var svcList = map[string]*apiobject.Service{}

type svcServiceHandler struct {
}

type svcPodHandler struct {
	//getURL() string
	//handleCreate(message string)
	//handleDelete(message string)
	//handleUpdate(message string)
}

/* ========== Start Service Handler ========== */

func (s svcServiceHandler) HandleCreate(message []byte) {
	svc := &apiobject.Service{}
	svc.UnMarshalJSON(message)

	// 1. allocate Cluster ip and update service
	svc.Spec.ClusterIP = allocateClusterIP()
	svcList[svc.Spec.ClusterIP] = svc
	utils.UpdateObject(svc, s.GetType(), svc.Data.Namespace, svc.Data.Name)

	// 2. traverse the pod list and create endpoints
	createEndpointsFromPodList(svc)

	log.Info("[svc controller] Create service. Cluster IP:", svc.Spec.ClusterIP)
}

func (s svcServiceHandler) HandleDelete(message []byte) {
	svc := &apiobject.Service{}
	svc.UnMarshalJSON(message)
	delete(svcList, svc.Spec.ClusterIP)

	// delete corresponding endpoints
	for _, edpt := range *svcToEndpoints[svc.Spec.ClusterIP] {
		utils.DeleteObject(utils.ENDPOINT, edpt.Data.Namespace, edpt.Data.Name)
	}
	delete(svcToEndpoints, svc.Spec.ClusterIP)

	log.Info("[svc controller] Delete service. Cluster IP:", svc.Spec.ClusterIP)
}

func (s svcServiceHandler) HandleUpdate(message []byte) {
	svc := &apiobject.Service{}
	svc.UnMarshalJSON(message)

	oldSvc, ok := svcList[svc.Spec.ClusterIP]
	if !ok {
		log.Info("[svc controller] Service not found. ClusterIP:", svc.Spec.ClusterIP)
		return
	}
	// check if the label changed. if so, delete old endpoints and add new ones
	if !utils.IsLabelEqual(oldSvc.Spec.Selector, svc.Spec.Selector) {
		for _, edpt := range *svcToEndpoints[svc.Spec.ClusterIP] {
			utils.DeleteObject(utils.ENDPOINT, edpt.Data.Namespace, edpt.Data.Name)
		}
		createEndpointsFromPodList(svc)
	}

	log.Info("[svc controller] Update service. Cluster IP:", svc.Spec.ClusterIP)
}

func (s svcServiceHandler) GetType() utils.ObjType {
	return utils.SERVICE
}

/* ========== Start Pod Handler ========== */

func (s svcPodHandler) HandleCreate(message []byte) {

}

func (s svcPodHandler) HandleDelete(message []byte) {

	pod := &apiobject.Pod{}
	pod.UnMarshalJSON(message)
	// delete corresponding endpoints
	for _, svc := range svcList {
		deleteEndpoints(svc, pod)
	}

}

func (s svcPodHandler) HandleUpdate(message []byte) {
	pod := &apiobject.Pod{}
	pod.UnMarshalJSON(message)

	for _, svc := range svcList {
		exist := isEndpointExist(svcToEndpoints[svc.Spec.ClusterIP], pod.Status.PodIp)
		fit := utils.IsLabelEqual(svc.Spec.Selector, pod.Data.Labels)
		if !exist && fit {
			createEndpoints(svcToEndpoints[svc.Spec.ClusterIP], svc, pod)
		} else if exist && !fit {
			deleteEndpoints(svc, pod)
		}
		// TODO: handle the change of POD IP
	}
}

func (s svcPodHandler) GetType() utils.ObjType {
	return utils.POD
}

/* ========== Util Function ========== */

func allocateClusterIP() string {
	for i, used := range IPMap {
		if !used {
			IPMap[i] = true
			return IPStart + strconv.Itoa(i)
		}
	}
	log.Fatal("[svc controller] Cluster IP used up!")
	return ""
}

func findDstPort(targetPort string, containers []apiobject.Container) int32 {
	for _, c := range containers {
		for _, p := range c.Ports {
			if p.Name == targetPort {
				return p.ContainerPort
			}
		}
	}
	log.Fatal("[svc controller] No Match for Target Port!")
	return 0
}

func createEndpoints(edptList *[]*apiobject.Endpoint, svc *apiobject.Service, pod *apiobject.Pod) {
	logInfo := "[svc controller] Create endpoints."

	for _, port := range svc.Spec.Ports {
		dstPort := findDstPort(port.TargetPort, pod.Spec.Containers)
		spec := apiobject.EndpointSpec{
			SvcIP:    svc.Spec.ClusterIP,
			SvcPort:  port.Port,
			DestIP:   pod.Status.PodIp,
			DestPort: dstPort,
		}
		edpt := &apiobject.Endpoint{
			Spec: spec,
			Data: apiobject.MetaData{
				Name:      svc.Data.Name + "-" + pod.Data.Name,
				Namespace: svc.Data.Namespace,
			},
		}
		utils.CreateObject(edpt, utils.ENDPOINT, svc.Data.Namespace)
		*edptList = append(*edptList, edpt)
		logInfo += fmt.Sprintf("srcIP:%s:%d, dstIP:%s:%d ; ", svc.Spec.ClusterIP, port.Port, pod.Status.PodIp, dstPort)
	}

	log.Info(logInfo)

}

func deleteEndpoints(svc *apiobject.Service, pod *apiobject.Pod) {
	logInfo := "[svc controller] Delete endpoints."

	edptList := svcToEndpoints[svc.Spec.ClusterIP]
	var newEdptList []*apiobject.Endpoint
	for key, edpt := range *edptList {
		if edpt.Spec.DestIP == pod.Status.PodIp {
			edpt := (*edptList)[key]
			utils.DeleteObject(utils.ENDPOINT, edpt.Data.Namespace, edpt.Data.Name)
			logInfo += fmt.Sprintf("srcIP:%s:%d, dstIP:%s:%d ; ", edpt.Spec.SvcIP, edpt.Spec.SvcPort, edpt.Spec.DestIP, edpt.Spec.DestPort)
		} else {
			newEdptList = append(newEdptList, edpt)
		}
	}
	svcToEndpoints[svc.Spec.ClusterIP] = &newEdptList

	log.Info(logInfo)
}

func isEndpointExist(edptList *[]*apiobject.Endpoint, podIP string) bool {
	for _, edpt := range *edptList {
		if edpt.Spec.DestIP == podIP {
			return true
		}
	}
	return false
}

func createEndpointsFromPodList(svc *apiobject.Service) {
	info := utils.GetObject(utils.POD, svc.Data.Namespace, "")
	podList := gjson.Parse(info).Array()
	var edptList []*apiobject.Endpoint
	for _, p := range podList {
		pod := &apiobject.Pod{}
		pod.UnMarshalJSON([]byte(p.String()))
		if pod.Data.ResourcesVersion != "delete" && utils.IsLabelEqual(svc.Spec.Selector, pod.Data.Labels) {
			createEndpoints(&edptList, svc, pod)
		}
	}
	svcToEndpoints[svc.Spec.ClusterIP] = &edptList
}
