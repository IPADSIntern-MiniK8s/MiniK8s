package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"minik8s/pkg/apiobject"
	kubeletPod "minik8s/pkg/kubelet/pod"
	"minik8s/utils"
	"net/http"
	"os"
	"time"
)

func register() {
	hostname, _ := os.Hostname()
	node := apiobject.Node{
		APIVersion: "v1",
		Kind:       "Node",
		Data: apiobject.MetaData{
			Name: hostname,
		},
		Spec: apiobject.NodeSpec{},
	}
	nodejson, _ := node.MarshalJSON()
	utils.SendJsonObject("POST", nodejson, "http://192.168.1.13:8080/api/v1/nodes")
}

func watchPod() {
	hostname, _ := os.Hostname()
	headers := http.Header{}
	headers.Set("X-Source", hostname)
	dialer := websocket.Dialer{}
	dialer.Jar = nil
	conn, _, err := dialer.Dial("ws://192.168.1.13:8080/api/v1/watch/pods", headers)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer conn.Close()
	var pod apiobject.Pod
	for {
		_, msgjson, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			continue
		}

		json.Unmarshal(msgjson, &pod)
		fmt.Println(pod.Status.Phase)
		switch pod.Status.Phase {
		case apiobject.Running:
			{
				success, ip := kubeletPod.CreatePod(pod)
				fmt.Println(success)
				if !success {
					continue
				}

				pod.Status.PodIp = ip
				pod.Status.Phase = apiobject.Succeeded
				break
			}
		//case apiobject.Terminating:{
		//		success:= kubeletPod.DeletePod(pod)
		//		if !success{
		//			continue
		//		}
		//		pod.Status.Phase = apiobject.Terminated
		//		break
		//	}
		default:
			continue
		}
		podjson, err := pod.MarshalJSON()
		if err != nil {
			fmt.Println(err)
			continue
		}
		utils.SendJsonObject("POST", podjson, fmt.Sprintf("http://%s/api/v1/namespaces/%s/pods/%s/update", utils.ApiServerIp, pod.Data.Namespace, pod.Data.Name))
	}
}

func main() {
	register()
	time.Sleep(time.Second * 5)
	watchPod()
}
