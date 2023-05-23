package eventfilter

import (
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/serverless/activator"
	"net/http"
)

func FunctionSync(target string) {
	// 建立WebSocket连接
	url := fmt.Sprintf("ws://%s/api/v1/watch/%s", config.ApiServerIp, target)
	log.Info("[FunctionSync] url: ", url)
	headers := http.Header{}
	headers.Set("X-Source", "function")
	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		fmt.Println("WebSocket connect fail", err)
		return
	} else {
		fmt.Println("WebSocket connect ")
	}
	defer conn.Close()

	// 不断地接收消息并处理
	log.Info("[FunctionSync] start to receive user message")
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

		op := gjson.Get(string(message), "status")
		// function trigger
		if !op.Exists() {
			FunctionTriggerHandler(message, conn)
			continue
		}
		switch op.String() {
		case "create":
			{
				FuntionCreateHandler(message, conn)
			}
		case "delete":
			{
				FunctionDeleteHandler(message, conn)
			}
		case "update":
			{
				FunctionUpdateHandler(message, conn)
			}
		}
	}
}

// TODO: need to add workflow later
func FuntionCreateHandler(message []byte, conn *websocket.Conn) {
	function := &apiobject.Function{}
	function.UnMarshalJSON(message)
	log.Info("[FuntionCreateHandler] function: ", function)

	// check the parameters
	if function.Name == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("function name is empty"))
		return
	}

	if function.Path == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("function path is empty"))
	}

	err := activator.InitFunc(function.Name, function.Path)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("function create success"))
	}
}

// FunctionTriggerHandler the trigger format: {"name": "function name", "params": "function params"}
func FunctionTriggerHandler(message []byte, conn *websocket.Conn) {
	nameField := gjson.Get(string(message), "name")
	if !nameField.Exists() {
		conn.WriteMessage(websocket.TextMessage, []byte("function name is empty"))
		return
	}

	name := nameField.String()
	paramsField := gjson.Get(string(message), "params")
	if !paramsField.Exists() {
		conn.WriteMessage(websocket.TextMessage, []byte("function params is empty"))
		return
	}

	params := paramsField.String()
	result, err := activator.TriggerFunc(name, []byte(params))
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}

	conn.WriteMessage(websocket.TextMessage, result)
}

func FunctionDeleteHandler(message []byte, conn *websocket.Conn) {
	function := &apiobject.Function{}
	function.UnMarshalJSON(message)
	log.Info("[FunctionDeleteHandler] function: ", function)

	// check the parameters
	if function.Name == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("function name is empty"))
		return
	}

	err := activator.DeleteFunc(function.Name)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("function delete success"))
	}

}

func FunctionUpdateHandler(message []byte, conn *websocket.Conn) {
	function := &apiobject.Function{}
	function.UnMarshalJSON(message)
	log.Info("[FunctionUpdateHandler] function: ", function)

	// delete the old function and create the new function
	err := activator.DeleteFunc(function.Name)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}

	err = activator.InitFunc(function.Name, function.Path)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("function update success"))
	}
}
