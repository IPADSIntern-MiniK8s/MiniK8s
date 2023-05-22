package kubescheduler

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	filter2 "minik8s/pkg/kubescheduler/filter"
	"minik8s/pkg/kubescheduler/policy"
	"minik8s/utils"
	"net/http"
)

type Config struct {
	ApiserverAddr string
	Policy        string
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

func Run(config Config) {
	// init scheduler and filter
	policyName := config.Policy
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

	// create websocket connection
	headers := http.Header{}
	headers.Set("X-Source", "scheduler")

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://"+config.ApiServerIp+"/api/v1/watch/pods", headers)
	if err != nil {
		log.Error("[Run] scheduler websocket connect fail")
		return
	}
	defer conn.Close()

	// create http client for ask api server
	httpMethod := "GET"
	httpUrl := "http://" + config.ApiServerIp + "/api/v1/nodes"

	// keep reading from websocket
	for {
		_, message, err := conn.ReadMessage()

		if len(message) == 0 {
			continue
		}

		if err != nil {
			log.Error("[Run] scheduler websocket read message fail")
			conn.WriteMessage(websocket.TextMessage, []byte{})
		}

		// parse message
		pod := &apiobject.Pod{}
		err = pod.UnMarshalJSON(message)
		if err != nil {
			log.Error("[Run] scheduler websocket unmarshal pod message fail")
			conn.WriteMessage(websocket.TextMessage, []byte{})
		}
		if pod == nil {
			conn.WriteMessage(websocket.TextMessage, []byte{})
		}

		// check whether pod is need to be scheduled
		if pod == nil || pod.Status.Phase != apiobject.Pending {
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
		conn.WriteMessage(websocket.TextMessage, jsonBytes)
	}

}
