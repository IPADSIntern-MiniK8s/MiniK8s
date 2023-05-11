package kubeproxy

/* 主要工作：
1. 监听service资源的创建。创建service
2. 监听service资源的删除。删除service
3. 监听endpoint的创建。设置dest规则。
4. 监听endpoint的删除。删除对应dest规则。
*/

import (
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeproxy/ipvs"
	"minik8s/utils"
	"strconv"
)

func Run() {
	ipvs.Init()
	//ipvs.TestConfig()
	var p proxyServiceHandler
	var e proxyEndpointHandler
	go utils.Sync(p)
	go utils.Sync(e)
	utils.WaitForever()
	//fmt.Println("end")

}

/* ========== Start Service Handler ========== */

type proxyServiceHandler struct {
}

func (p proxyServiceHandler) HandleCreate(message []byte) {

}

func (p proxyServiceHandler) HandleDelete(message []byte) {
	svc := &apiobject.Service{}
	svc.UnMarshalJSON(message)

	for _, p := range svc.Spec.Ports {
		key := svc.Status.ClusterIP + ":" + strconv.Itoa(int(p.Port))
		ipvs.DeleteService(key)
	}

}

func (p proxyServiceHandler) HandleUpdate(message []byte) {
	svc := &apiobject.Service{}
	svc.UnMarshalJSON(message)

	for _, p := range svc.Spec.Ports {
		ipvs.AddService(svc.Status.ClusterIP, uint16(p.Port))
	}

}

func (p proxyServiceHandler) GetType() utils.ObjType {
	return utils.SERVICE
}

/* ========== Start Endpoint Handler ========== */

type proxyEndpointHandler struct {
}

func (e proxyEndpointHandler) HandleCreate(message []byte) {
	edpt := &apiobject.Endpoint{}
	edpt.UnMarshalJSON(message)

	key := edpt.Spec.SvcIP + ":" + strconv.Itoa(int(edpt.Spec.SvcPort))
	ipvs.AddEndpoint(key, edpt.Spec.DestIP, uint16(edpt.Spec.DestPort))
}

func (e proxyEndpointHandler) HandleDelete(message []byte) {
	edpt := &apiobject.Endpoint{}
	edpt.UnMarshalJSON(message)

	svcKey := edpt.Spec.SvcIP + ":" + strconv.Itoa(int(edpt.Spec.SvcPort))
	dstKey := edpt.Spec.DestIP + ":" + strconv.Itoa(int(edpt.Spec.DestPort))
	ipvs.DeleteEndpoint(svcKey, dstKey)
}

func (e proxyEndpointHandler) HandleUpdate(message []byte) {

}

func (e proxyEndpointHandler) GetType() utils.ObjType {
	return utils.ENDPOINT
}

//func HandleServiceChange(message string) {
//	// traverse the service list, add the new service
//	serviceList := gjson.Get(message, "").Array()
//	for _, svc := range serviceList {
//		clusterIP := gjson.Get(svc.String(), "spec.clusterIP").String()
//		ports := gjson.Get(svc.String(), "spec.ports").Array()
//		for _, port := range ports {
//			serviceIP := clusterIP + ":" + port.String()
//			if node, ok := ipvs.Services[serviceIP]; !ok {
//				ipvs.AddService(clusterIP, uint16(port.Int()))
//			} else {
//				node.Visited = true
//			}
//		}
//	}
//
//	// traverse the service list in proxy, delete the service not visited
//	for k, node := range ipvs.Services {
//		if node.Visited == false {
//			ipvs.DeleteService(k, node)
//		} else {
//			node.Visited = false
//		}
//	}
//}
//
//func HandleEndpointChange(message string) {
//	eptList := gjson.Get(message, "").Array()
//	for _, epts := range eptList {
//		clusterIP := gjson.Get(epts.String(), "clusterIP").String()
//		port := gjson.Get(epts.String(), "port").String()
//		serviceIP := clusterIP + port
//		if svc, ok := ipvs.Services[serviceIP]; ok {
//			dests := gjson.Get(epts.String(), "subsets").Array()
//			// traverse the endpoints list, add the new endpoints
//			for _, dest := range dests {
//				ip := gjson.Get(dest.String(), "IP").String()
//				port := gjson.Get(dest.String(), "Port").Int()
//				podIP := ip + ":" + string(port)
//				if edpNode, ok := svc.Endpoints[podIP]; !ok {
//					ipvs.AddEndpoint(svc, ip, uint16(port))
//				} else {
//					edpNode.Visited = true
//				}
//			}
//			// traverse the endpoint list in proxy, delete the endpoint not visited
//			for k, node := range svc.Endpoints {
//				if node.Visited == false {
//					ipvs.DeleteEndpoint(svc, node.Endpoint, k)
//				} else {
//					node.Visited = false
//				}
//			}
//
//		}
//	}
//
//}

func syncRunner() {

}
