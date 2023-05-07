package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"minik8s/pkg/apiobject"
	kubeletPod "minik8s/pkg/kubelet/pod"
	"minik8s/utils"
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
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial("ws://192.168.1.13:8080/api/v1/watch/"+hostname+"/pods", nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	var pod apiobject.Pod
	for {
		_, msgjson, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			continue
		}

		json.Unmarshal(msgjson, &pod)
		if pod.Status.Phase == "Ready" {
			success := kubeletPod.CreatePod(pod)
			fmt.Println(success)
		}

	}
}

func main() {
	register()
	time.Sleep(time.Second * 5)
	watchPod()
}
