package kubeproxy

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"minik8s/pkg/kubeproxy/ipvs"
	"minik8s/utils"
)

func Run() {
	ipvs.Init()
	//ipvs.TestConfig()
	go sync(SERVICE)
	go sync(ENDPOINT)
	//fmt.Println("end")
	//runLoop()

}

type ObjType string

const (
	SERVICE  ObjType = "services"
	ENDPOINT ObjType = "endpoints"
)

var handleFunc = map[ObjType]func(string){
	SERVICE:  handleServiceChange,
	ENDPOINT: handleEndpointChange,
}

func sync(objType ObjType) {
	// 建立WebSocket连接
	url := fmt.Sprintf("ws://%s/api/v1/%s", utils.ApiServerIp, objType)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("WebSocket连接失败：", err)
		return
	} else {
		fmt.Println("WebSocket连接成功")
	}

	// 不断地接收消息并处理
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("读取消息失败：", err)
			return
		}
		fmt.Printf("%s\n", message)

		handleFunc[objType](string(message))
	}
}

func handleServiceChange(message string) {
	// traverse the service list, add the new service
	serviceList := gjson.Get(message, "").Array()
	for _, svc := range serviceList {
		clusterIP := gjson.Get(svc.String(), "spec.clusterIP").String()
		ports := gjson.Get(svc.String(), "spec.ports").Array()
		for _, port := range ports {
			serviceIP := clusterIP + ":" + port.String()
			if node, ok := ipvs.Services[serviceIP]; !ok {
				ipvs.AddService(clusterIP, uint16(port.Int()))
			} else {
				node.Visited = true
			}
		}
	}

	// traverse the service list in proxy, delete the service not visited
	for k, node := range ipvs.Services {
		if node.Visited == false {
			ipvs.DeleteService(k, node)
		} else {
			node.Visited = false
		}
	}
}

func handleEndpointChange(message string) {
	eptList := gjson.Get(message, "").Array()
	for _, epts := range eptList {
		clusterIP := gjson.Get(epts.String(), "clusterIP").String()
		port := gjson.Get(epts.String(), "port").String()
		serviceIP := clusterIP + port
		if svc, ok := ipvs.Services[serviceIP]; ok {
			dests := gjson.Get(epts.String(), "subsets").Array()
			// traverse the endpoints list, add the new endpoints
			for _, dest := range dests {
				ip := gjson.Get(dest.String(), "IP").String()
				port := gjson.Get(dest.String(), "Port").Int()
				podIP := ip + ":" + string(port)
				if edpNode, ok := svc.Endpoints[podIP]; !ok {
					ipvs.AddEndpoint(svc, ip, uint16(port))
				} else {
					edpNode.Visited = true
				}
			}
			// traverse the endpoint list in proxy, delete the endpoint not visited
			for k, node := range svc.Endpoints {
				if node.Visited == false {
					ipvs.DeleteEndpoint(svc, node.Endpoint, k)
				} else {
					node.Visited = false
				}
			}

		}
	}

}

func syncRunner() {

}
