package kubelet

import (
	"fmt"
	"github.com/gorilla/websocket"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubelet/metricsserver"
	kubeletPod "minik8s/pkg/kubelet/pod"
	"minik8s/utils"
	"net/http"
	"os"
	"time"
	"github.com/tidwall/gjson"
)

type Kubelet struct {
	ApiserverAddr string            //communicate with server
	FlannelSubnet string            //service
	IP            string            //host ip
	Labels        map[string]string // for nodeSelector
	ListenAddr    string            //MetricsServer listen for auto-scaling
	CPU		string
	Memory		string
	Server        *metricsserver.MetricsServer
}

func NewKubelet(config Config) *Kubelet {
	return &Kubelet{
		ApiserverAddr: config.ApiserverAddr,
		FlannelSubnet: config.FlannelSubnet,
		IP:            config.IP,
		Labels:        config.Labels,
		ListenAddr:    config.ListenAddr,
		CPU:		config.CPU,
		Memory:		config.Memory,
		Server:        metricsserver.NewMetricsServer(),
	}
}

func (kl *Kubelet) register() {
	hostname, _ := os.Hostname()
	node := apiobject.Node{
		APIVersion: "v1",
		Kind:       "Node",
		Data: apiobject.MetaData{
			Name:   hostname,
			Labels: kl.Labels,
		},
		Spec: apiobject.NodeSpec{
			Unschedulable: false,
			PodCIDR:       kl.FlannelSubnet,
		},
		Status: apiobject.NodeStatus{
			Capability:map[string]string{
				"cpu":kl.CPU,
				"memory":kl.Memory,
			},
			Addresses: []apiobject.Address{
				{
					Type:    "InternalIP",
					Address: kl.IP,
				},
			},
		},
	}
	nodejson, _ := node.MarshalJSON()
	utils.SendJsonObject("POST", nodejson, fmt.Sprintf("http://%s/api/v1/nodes", kl.ApiserverAddr))
}

func (kl *Kubelet) watchPod() {
	hostname, _ := os.Hostname()
	headers := http.Header{}
	headers.Set("X-Source", hostname)
	dialer := websocket.Dialer{}
	dialer.Jar = nil
	conn, _, err := dialer.Dial(fmt.Sprintf("ws://%s/api/v1/watch/pods", kl.ApiserverAddr), headers)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer conn.Close()
	for {
		pod := &apiobject.Pod{}
		_, msgjson, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			continue
		}
		pod.UnMarshalJSON(msgjson)
		if pod.Status.HostIp != kl.IP {
			continue
		}
		switch pod.Status.Phase {
		case apiobject.Scheduled:
			{
				success, ip := kubeletPod.CreatePod(*pod, kl.ApiserverAddr)
				fmt.Println(success)
				if !success {
					pod.Status.Phase = apiobject.Failed
					utils.UpdateObject(pod, config.POD, pod.Data.Namespace, pod.Data.Name)
					continue
				}

				pod.Status.PodIp = ip
				pod.Status.Phase = apiobject.Running
				break
			}
		case apiobject.Terminating:
			{
				success := kubeletPod.DeletePod(*pod)
				if !success {
					continue
				}
				pod.Status.Phase = apiobject.Deleted
				break
			}
		default:
			continue
		}
		utils.UpdateObject(pod, config.POD, pod.Data.Namespace, pod.Data.Name)
		//time.Sleep(time.Millisecond * 200)
		//podjson, err := pod.MarshalJSON()
		//if err != nil {
		//	fmt.Println(err)
		//	continue
		//}
		//utils.SendJsonObject("POST", podjson, fmt.Sprintf("http://%s/api/v1/namespaces/%s/pods/%s/update", kl.ApiserverAddr, pod.Data.Namespace, pod.Data.Name))
	}

}


func(kl *Kubelet) watchContainersStatus(){
	for{
		time.Sleep(time.Second*10)

		url := fmt.Sprintf("http://%s/api/v1/pods", kl.ApiserverAddr)
		hostname, _ := os.Hostname()
		info, err := utils.SendRequestWithHb("GET", nil, url, hostname)
		if err != nil {
			fmt.Println(err)
			continue
		}
		podList := gjson.Parse(info).Array()
		for _, p := range podList {
			pod := &apiobject.Pod{}
			pod.UnMarshalJSON([]byte(p.String()))

			phase,stopped := kubeletPod.GetPodPhase(*pod)
			if stopped{
				fmt.Println(pod.Data.Name,phase)
				pod.Status.Phase = phase
				utils.UpdateObject(pod, config.POD, pod.Data.Namespace, pod.Data.Name)
			}
		}
	}
}
