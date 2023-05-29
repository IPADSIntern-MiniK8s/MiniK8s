package eventfilter

import (
	"fmt"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/serverless/workflow"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func WorkFlowSync(target string) {
	for {
		err := workflowConnect(target)
		if err != nil {
			log.Error("[WorkFlowSync] WebSocket connect fail: ", err)
		}
		time.Sleep(5 * time.Second) // wait 5 seconds to reconnect
	}
}


func workflowConnect(target string) error {
	// establish WebSocket connection
	url := fmt.Sprintf("ws://%s/api/v1/watch/%s", config.ApiServerIp, target)
	headers := http.Header{}
	headers.Set("X-Source", "workflows")
	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		fmt.Println("WebSocket connect fail", err)
		return err
	} else {
		fmt.Println("WebSocket connect ")
	}
	defer conn.Close()

	// continue to receive messages and process
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read from websocket fail: ", err)
			return err
		}
		if len(message) == 0 {
			continue
		}
		fmt.Printf("[client %s] %s\n", target, message)

		workFlow := gjson.Get(string(message), "workflow")
		if !workFlow.Exists() {
			conn.WriteMessage(websocket.TextMessage, []byte("execute: " + "the workFlow is not exist"))
		}
		workFlowStr := workFlow.String()
		log.Info("[WorkFlowSync] workFlow: ", workFlowStr)

		params := gjson.Get(string(message), "params")
		if !params.Exists() {
			conn.WriteMessage(websocket.TextMessage, []byte("execute: " + "the params is not exist"))
		}
		paramsStr := params.String()

		go WorkFlowTriggerHandler([]byte(workFlowStr), []byte(paramsStr), conn)
		// WorkFlowTriggerHandler([]byte(workFlowStr), []byte(paramsStr), conn)
	}
}

func WorkFlowTriggerHandler(workFlow []byte, paramsStr []byte, conn *websocket.Conn) {
	// parse the workFlow
	currentWorkFlow := &apiobject.WorkFlow{}
	err := currentWorkFlow.UnMarshalJSON(workFlow)
	if err != nil {
		log.Error("[WorkFlowTriggerHandler] unmarshal workFlow error: ", err)
		conn.WriteMessage(websocket.TextMessage, []byte("execute: " + "unmarshal workFlow error"))
	}
	result, err := workflow.ExecuteWorkFlow(currentWorkFlow, paramsStr)
	if err != nil {
		log.Error("[WorkFlowTriggerHandler] execute workFlow error: ", err)
		conn.WriteMessage(websocket.TextMessage, []byte("execute: " + err.Error()))
	}
	conn.WriteMessage(websocket.TextMessage, []byte("execute: " + string(result)))
}

