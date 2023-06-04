package kubelet

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubelet/metricsserver"
	kubeletPod "minik8s/pkg/kubelet/pod"
	kubeletUtils "minik8s/pkg/kubelet/utils"
	"minik8s/utils"
	"net/http"
	"os"
	"sync"
	"time"
)

type Kubelet struct {
	ApiserverAddr string            //communicate with server
	FlannelSubnet string            //service
	IP            string            //host ip
	Labels        map[string]string // for nodeSelector
	ListenAddr    string            //MetricsServer listen for auto-scaling
	CPU           string
	Memory        string
	Server        *metricsserver.MetricsServer
	Locks         sync.Map
}

func NewKubelet(config Config) *Kubelet {
	return &Kubelet{
		ApiserverAddr: config.ApiserverAddr,
		FlannelSubnet: config.FlannelSubnet,
		IP:            config.IP,
		Labels:        config.Labels,
		ListenAddr:    config.ListenAddr,
		CPU:           config.CPU,
		Memory:        config.Memory,
		Server:        metricsserver.NewMetricsServer(),
		Locks:         sync.Map{},
	}
}

func (kl *Kubelet) register() bool {
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
			Capability: map[string]string{
				"cpu":    kl.CPU,
				"memory": kl.Memory,
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
	return utils.SendJsonObject("POST", nodejson, fmt.Sprintf("http://%s/api/v1/nodes", kl.ApiserverAddr))
}

func (kl *Kubelet) watchPod() {
	hostname, _ := os.Hostname()
	headers := http.Header{}
	headers.Set("X-Source", hostname)
	dialer := websocket.Dialer{}
	dialer.Jar = nil
	conn, _, err := dialer.Dial(fmt.Sprintf("ws://%s/api/v1/watch/pods", kl.ApiserverAddr), headers)
	if err != nil {
		//fmt.Println(err)
		//panic(err)
		return
	}
	defer conn.Close()
	for {
		pod := &apiobject.Pod{}
		_, msgjson, err := conn.ReadMessage()
		fmt.Println("received message from apiserver")
		if err != nil {
			//fmt.Println(err)
			//continue
			return
		}
		pod.UnMarshalJSON(msgjson)
		if pod.Status.HostIp != kl.IP {
			continue
		}
		fmt.Println(*pod)
		switch pod.Status.Phase {
		case apiobject.Scheduled:
			{
				go func(p apiobject.Pod){
					kubeletUtils.Lock(pod.Data.Namespace, pod.Data.Name)
					success, ip := kubeletPod.CreatePod(*pod, kl.ApiserverAddr)
					kubeletUtils.UnLock(pod.Data.Namespace, pod.Data.Name)
					fmt.Println(success)
					if success {
						pod.Status.PodIp = ip
						pod.Status.Phase = apiobject.Running
						utils.UpdateObject(pod, config.POD, pod.Data.Namespace, pod.Data.Name)
					}else{
						pod.Status.Phase = apiobject.Failed
						utils.UpdateObject(pod, config.POD, pod.Data.Namespace, pod.Data.Name)
					}
				}(*pod)

			}
		case apiobject.Terminating:
			{
				go func(p apiobject.Pod){
					kubeletUtils.Lock(pod.Data.Namespace, pod.Data.Name)
					success := kubeletPod.DeletePod(*pod)
					kubeletUtils.UnLock(pod.Data.Namespace, pod.Data.Name)
					if !success {
						pod.Status.Phase = apiobject.Deleted
						utils.UpdateObject(pod, config.POD, pod.Data.Namespace, pod.Data.Name)
					}
				}(*pod)
			}
		default:
			continue
		}
	}

}

func (kl *Kubelet) watchContainersStatus() {
	for {
		time.Sleep(time.Second * 10)

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
			kubeletUtils.RLock(pod.Data.Namespace, pod.Data.Name)
			phase, stopped := kubeletPod.GetPodPhase(*pod)
			if stopped {
				fmt.Println(pod.Data.Name, phase)
				pod.Status.Phase = phase
				utils.UpdateObject(pod, config.POD, pod.Data.Namespace, pod.Data.Name)
			}
			kubeletUtils.RUnLock(pod.Data.Namespace, pod.Data.Name)

		}
	}
}
