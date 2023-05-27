package kubescheduler

import (
	"minik8s/config"
	"minik8s/pkg/apiobject"
	filter2 "minik8s/pkg/kubescheduler/filter"
	"minik8s/pkg/kubescheduler/policy"
	"minik8s/utils"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Policy string
}

// generate the new pointer slice
func toPointerSlice(slice []apiobject.Node) []*apiobject.Node {
	result := make([]*apiobject.Node, len(slice))
	for i, v := range slice {
		// create a new pointer which pointer to the node
		node := v
		result[i] = &node
	}
	return result
}

func toValueSlice(slice []*apiobject.Node) []apiobject.Node {
	result := make([]apiobject.Node, len(slice))
	for i, v := range slice {
		result[i] = *v
	}
	return result
}

func connect(scheduler policy.Scheduler) error {
	// create websocket connection
	headers := http.Header{}
	headers.Set("X-Source", "scheduler")

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://"+config.ApiServerIp+"/api/v1/watch/pods", headers)
	if err != nil {
		log.Error("[Run] scheduler websocket connect fail")
		time.Sleep(5 * time.Second) // wait 5 seconds to reconnect
		return err
	}
	defer conn.Close()

	// create http client for ask api server
	httpMethod := "GET"
	httpUrl := "http://" + config.ApiServerIp + "/api/v1/nodes"

	// keep reading from websocket
	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Error("[Run] scheduler websocket read message fail")
			return err 
		}

		if len(message) == 0 {
			continue
		}

		// parse message
		pod := &apiobject.Pod{}
		err = pod.UnMarshalJSON(message)
		if err != nil {
			log.Error("[Run] scheduler websocket unmarshal pod message fail")
			conn.WriteMessage(websocket.TextMessage, []byte{})
		}
		if pod == nil {
			log.Error("[Run] scheduler websocket pod is nil")
			conn.WriteMessage(websocket.TextMessage, []byte{})
		}

		// check whether pod is need to be scheduled
		if pod.Status.Phase != apiobject.Pending {
			log.Error("[Run] scheduler websocket pod is nil or pod is not pending, the pod phase is: ", pod.Status.Phase)
			conn.WriteMessage(websocket.TextMessage, []byte{})
		}

		// get the nodes that pod will be scheduled to
		response, err := utils.SendRequest(httpMethod, []byte{}, httpUrl)
		if err != nil {
			log.Error("[Run] scheduler http request fail, the error message is: ", err)
			conn.WriteMessage(websocket.TextMessage, []byte{})
		}

		// get the nodes
		node := &apiobject.Node{}
		nodes, err := node.UnMarshalJSONList([]byte(response))
		if err != nil {
			log.Error("[Run] scheduler unmarshal nodes fail, the error message is: ", err)
			conn.WriteMessage(websocket.TextMessage, []byte{})
		}

		// schedule pod
		nodeList := toPointerSlice(nodes)
		nodeForCandidates := scheduler.Schedule(pod, nodeList)
		nodeCandidates := toValueSlice(nodeForCandidates)

		// marshal the nodes that pod will be scheduled to and send to api server
		jsonBytes, err := apiobject.MarshalJSONList(nodeCandidates)
		if err != nil {
			log.Error("[Run] scheduler marshal nodes fail, the error message is: ", err)
		} else {
			log.Info("[Run] the node candidate  count is: ", len(nodeCandidates))
		}
		conn.WriteMessage(websocket.TextMessage, jsonBytes)
	}
}

func Run(c Config) {
	// init scheduler and filter
	policyName := c.Policy
	var filter filter2.TemplateFilter
	concreteFilter := filter2.ConfigFilter{Name: "ConfigFilter"}
	filter = concreteFilter
	var scheduler policy.Scheduler

	if policyName == "default" || policyName == "resource" {
		concreteScheduler := policy.NewResourceScheduler(filter)
		scheduler = concreteScheduler
	} else if policyName == "frequency" {
		concreteScheduler := policy.NewLeastRequestScheduler(filter)
		scheduler = concreteScheduler
	}

	

	for {
		err := connect(scheduler)
		if err != nil {
			log.Error("[Run] scheduler connect fail, the error message is: ", err)
		}
		time.Sleep(5 * time.Second)
	}
	
}
