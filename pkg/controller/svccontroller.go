package controller

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"minik8s/pkg/apiobject"
	"minik8s/utils"
)

/* 主要工作：
1. 监听service资源的创建。一旦service资源创建，为其分配唯一的cluster ip。
2. 遍历pod列表，找到符合selector条件的pod，记录。创建endpoint。
3. 监听pod列表，一旦pod删除。删除endpoint。
4. 监听service资源的删除。删除对应endpoint。
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
	info := utils.GetObjects(utils.POD)
	podList := gjson.Get(info, "").Array()
	var edptList []*apiobject.Endpoint
	for _, p := range podList {
		pod := &apiobject.Pod{}
		pod.UnMarshalJSON([]byte(p.String()))
		if utils.IsPodFitSelector(svc.Spec.Selector, pod.Data.Labels) {
			createEndpoints(&edptList, svc, pod)
		}
	}
	svcToEndpoints[svc.Spec.ClusterIP] = &edptList

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

	log.Info("[svc controller] Delete service. Cluster IP:", svc.Spec.ClusterIP)
}

func (s svcServiceHandler) HandleUpdate(message []byte) {
	svc := &apiobject.Service{}
	svc.UnMarshalJSON(message)

	log.Info("[svc controller] Update service. Cluster IP:", svc.Spec.ClusterIP)
}

func (s svcServiceHandler) GetType() utils.ObjType {
	return utils.SERVICE
}

/* ========== Start Pod Handler ========== */

func (s svcPodHandler) HandleCreate(message []byte) {

}

func (s svcPodHandler) HandleDelete(message []byte) {
	logInfo := "[svc controller] Delete endpoints.\n"

	pod := &apiobject.Pod{}
	pod.UnMarshalJSON(message)
	// delete corresponding endpoints
	for _, svc := range svcList {
		edptList := svcToEndpoints[svc.Spec.ClusterIP]
		var newEdptList []*apiobject.Endpoint
		for key, edpt := range *edptList {
			if edpt.Spec.DestIP == pod.Status.PodIp {
				edpt := (*edptList)[key]
				utils.DeleteObject(utils.ENDPOINT, edpt.Data.Namespace, edpt.Data.Name)
				logInfo += fmt.Sprintf("srcIP:%s:%d, dstIP:%s:%d\n", edpt.Spec.SvcIP, edpt.Spec.SvcPort, edpt.Spec.DestIP, edpt.Spec.DestPort)
			} else {
				newEdptList = append(newEdptList, edpt)
			}
		}
		svcToEndpoints[svc.Spec.ClusterIP] = &newEdptList
	}

	log.Info(logInfo)
}

func (s svcPodHandler) HandleUpdate(message []byte) {
	pod := &apiobject.Pod{}
	pod.UnMarshalJSON(message)

	for _, svc := range svcList {
		if utils.IsPodFitSelector(svc.Spec.Selector, pod.Data.Labels) {
			createEndpoints(svcToEndpoints[svc.Spec.ClusterIP], svc, pod)
		}
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
			return IPStart + string(i)
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
	logInfo := "[svc controller] Create endpoints.\n"

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
				Name:             svc.Data.Name + "-" + pod.Data.Name,
				Namespace:        svc.Data.Namespace,
				ResourcesVersion: "CREATE",
			},
		}
		utils.CreateObject(edpt, utils.ENDPOINT, svc.Data.Namespace)
		*edptList = append(*edptList, edpt)
		logInfo += fmt.Sprintf("srcIP:%s:%d, dstIP:%s:%d\n", svc.Spec.ClusterIP, port.Port, pod.Status.PodIp, dstPort)
	}

	log.Info(logInfo)
}
