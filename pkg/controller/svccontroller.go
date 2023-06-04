package controller

import (
	"fmt"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/utils"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
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
	svc.Status.ClusterIP = allocateClusterIP()
	svcList[svc.Status.ClusterIP] = svc
	utils.UpdateObject(svc, s.GetType(), svc.Data.Namespace, svc.Data.Name)

	// 2. traverse the pod list and create endpoints
	createEndpointsFromPodList(svc)

	log.Info("[svc controller] Create service. Cluster IP:", svc.Status.ClusterIP)
}

func (s svcServiceHandler) HandleDelete(message []byte) {
	svc := &apiobject.Service{}
	svc.UnMarshalJSON(message)
	delete(svcList, svc.Status.ClusterIP)
	index := strings.SplitN(svc.Status.ClusterIP, ",", -1)
	indexLast, _ := strconv.Atoi(index[len(index)-1])
	print(indexLast)
	IPMap[indexLast] = false

	// delete corresponding endpoints
	for _, edpt := range *svcToEndpoints[svc.Status.ClusterIP] {
		utils.DeleteObject(config.ENDPOINT, edpt.Data.Namespace, edpt.Data.Name)
	}
	delete(svcToEndpoints, svc.Status.ClusterIP)

	log.Info("[svc controller] Delete service. Cluster IP:", svc.Status.ClusterIP)
}

func (s svcServiceHandler) HandleUpdate(message []byte) {
	svc := &apiobject.Service{}
	svc.UnMarshalJSON(message)

	oldSvc, ok := svcList[svc.Status.ClusterIP]
	if !ok {
		log.Info("[svc controller] Service not found. ClusterIP:", svc.Status.ClusterIP, "\n")
		return
	}
	// check if the label changed. if so, delete old endpoints and add new ones
	if !utils.IsLabelEqual(oldSvc.Spec.Selector, svc.Spec.Selector) {
		for _, edpt := range *svcToEndpoints[svc.Status.ClusterIP] {
			utils.DeleteObject(config.ENDPOINT, edpt.Data.Namespace, edpt.Data.Name)
		}
		createEndpointsFromPodList(svc)
	}

	svcList[svc.Status.ClusterIP] = svc
	log.Info("[svc controller] Update service. Cluster IP:", svc.Status.ClusterIP)
}

func (s svcServiceHandler) GetType() config.ObjType {
	return config.SERVICE
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
		exist := isEndpointExist(svcToEndpoints[svc.Status.ClusterIP], pod.Status.PodIp)
		fit := utils.IsLabelEqual(svc.Spec.Selector, pod.Data.Labels)
		if !exist && fit {
			createEndpoints(svcToEndpoints[svc.Status.ClusterIP], svc, pod)
		} else if exist && !fit {
			deleteEndpoints(svc, pod)
		}
		// TODO: handle the change of POD IP
	}
}

func (s svcPodHandler) GetType() config.ObjType {
	return config.POD
}

/* ========== Util Function ========== */

func allocateClusterIP() string {
	for i, used := range IPMap {
		if i != 0 && !used {
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
			SvcIP:    svc.Status.ClusterIP,
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
		utils.CreateObject(edpt, config.ENDPOINT, svc.Data.Namespace)
		*edptList = append(*edptList, edpt)
		logInfo += fmt.Sprintf("srcIP:%s:%d, dstIP:%s:%d ; ", svc.Status.ClusterIP, port.Port, pod.Status.PodIp, dstPort)
	}

	log.Info(logInfo)

}

func deleteEndpoints(svc *apiobject.Service, pod *apiobject.Pod) {
	logInfo := "[svc controller] Delete endpoints."

	edptList := svcToEndpoints[svc.Status.ClusterIP]
	var newEdptList []*apiobject.Endpoint
	for key, edpt := range *edptList {
		if edpt.Spec.DestIP == pod.Status.PodIp {
			edpt := (*edptList)[key]
			utils.DeleteObject(config.ENDPOINT, edpt.Data.Namespace, edpt.Data.Name)
			logInfo += fmt.Sprintf("srcIP:%s:%d, dstIP:%s:%d ; ", edpt.Spec.SvcIP, edpt.Spec.SvcPort, edpt.Spec.DestIP, edpt.Spec.DestPort)
		} else {
			newEdptList = append(newEdptList, edpt)
		}
	}
	svcToEndpoints[svc.Status.ClusterIP] = &newEdptList

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
	info := utils.GetObject(config.POD, svc.Data.Namespace, "")
	podList := gjson.Parse(info).Array()
	var edptList []*apiobject.Endpoint
	for _, p := range podList {
		pod := &apiobject.Pod{}
		pod.UnMarshalJSON([]byte(p.String()))
		if pod.Data.ResourcesVersion != "delete" && utils.IsLabelEqual(svc.Spec.Selector, pod.Data.Labels) {
			createEndpoints(&edptList, svc, pod)
		}
	}
	svcToEndpoints[svc.Status.ClusterIP] = &edptList
}
