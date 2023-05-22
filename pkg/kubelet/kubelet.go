package kubelet

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubelet/metricsserver"
	kubeletPod "minik8s/pkg/kubelet/pod"
	"minik8s/utils"
	"net/http"
	"os"
)

type Kubelet struct {
	ApiserverAddr string            //communicate with server
	FlannelSubnet string            //service
	IP            string            //host ip
	Labels        map[string]string // for nodeSelector
	ListenAddr    string            //MetricsServer listen for auto-scaling
	Server        *metricsserver.MetricsServer
}

func NewKubelet(config Config) *Kubelet {
	return &Kubelet{
		ApiserverAddr: config.ApiserverAddr,
		FlannelSubnet: config.FlannelSubnet,
		IP:            config.IP,
		Labels:        config.Labels,
		ListenAddr:    config.ListenAddr,
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
	var pod apiobject.Pod
	for {
		_, msgjson, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			continue
		}

		json.Unmarshal(msgjson, &pod)
		fmt.Println(pod.Status.Phase)
		if pod.Status.HostIp != kl.IP {
			continue
		}
		switch pod.Status.Phase {
		case apiobject.Running:
			{
				success, ip := kubeletPod.CreatePod(pod, kl.ApiserverAddr)
				fmt.Println(success)
				if !success {
					continue
				}

				pod.Status.PodIp = ip
				pod.Status.Phase = apiobject.Succeeded
				break
			}
		case apiobject.Terminating:
			{
				success := kubeletPod.DeletePod(pod)
				if !success {
					continue
				}
				pod.Status.Phase = apiobject.Deleted
				break
			}
		default:
			continue
		}
		//utils.UpdateObject(&pod, utils.POD, pod.Data.Namespace, pod.Data.Name)
		//time.Sleep(time.Millisecond * 200)
		podjson, err := pod.MarshalJSON()
		if err != nil {
			fmt.Println(err)
			continue
		}
		utils.SendJsonObject("POST", podjson, fmt.Sprintf("http://%s/api/v1/namespaces/%s/pods/%s/update", kl.ApiserverAddr, pod.Data.Namespace, pod.Data.Name))
	}
}
