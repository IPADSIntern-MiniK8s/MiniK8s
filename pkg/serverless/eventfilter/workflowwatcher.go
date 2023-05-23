package eventfilter

import (
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/serverless/workflow"
)

func WorkFlowSync(target string) {
	// establish WebSocket connection
	url := fmt.Sprintf("ws://%s/api/v1/watch/%s", config.ApiServerIp, target)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("WebSocket connect fail", err)
		return
	} else {
		fmt.Println("WebSocket connect ")
	}
	defer conn.Close()

	// continue to receive messages and process
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read from websocket fail: ", err)
			return
		}
		if len(message) == 0 {
			continue
		}
		fmt.Printf("[client %s] %s\n", target, message)
		
		workFlow := gjson.Get(string(message), "workFlow")
		if !workFlow.Exists() {
			conn.WriteMessage(websocket.TextMessage, []byte("the workFlow is not exist"))
		}
		workFlowStr := workFlow.String()


		params := gjson.Get(string(message), "params")
		if !params.Exists() {
			conn.WriteMessage(websocket.TextMessage, []byte("the params is not exist"))
		}
		paramsStr := params.String()

		go WorkFlowTriggerHandler([]byte(workFlowStr), []byte(paramsStr), conn)
	}
} 


func WorkFlowTriggerHandler(workFlow []byte, paramsStr []byte, conn *websocket.Conn) {
	// parse the workFlow
	currentWorkFlow := &apiobject.WorkFlow{}
	err := currentWorkFlow.UnMarshalJSON(workFlow)
	if err != nil {
		log.Error("[WorkFlowTriggerHandler] unmarshal workFlow error: ", err)
		conn.WriteMessage(websocket.TextMessage, []byte("unmarshal workFlow error"))
	}
	result, err := workflow.ExecuteWorkFlow(currentWorkFlow, paramsStr)
	if err != nil {
		log.Error("[WorkFlowTriggerHandler] execute workFlow error: ", err)
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	}
	conn.WriteMessage(websocket.TextMessage, result)
}
	