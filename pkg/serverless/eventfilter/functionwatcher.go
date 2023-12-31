package eventfilter

import (
	"fmt"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/serverless/activator"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func FunctionSync(target string) {
	// establish websocket connection
	for {
		err := connect(target)
		if err != nil {
			log.Error("[FunctionSync] WebSocket connect fail: ", err)
		}
		time.Sleep(5 * time.Second) // wait 5 seconds to reconnect
	}
}

func connect(target string) error {
	// establish websocket connection
	url := fmt.Sprintf("ws://%s/api/v1/watch/%s", config.ApiServerIp, target)
	log.Info("[FunctionSync] url: ", url)
	headers := http.Header{}
	headers.Set("X-Source", "function")
	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		fmt.Println("WebSocket connect fail", err)
		return err
	} else {
		fmt.Println("WebSocket connect ")
	}
	defer conn.Close()

	// continuously receive message
	log.Info("[FunctionSync] start to receive user message")
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

		op := gjson.Get(string(message), "status")
		// function trigger
		if !op.Exists() {
			go FunctionTriggerHandler(message, conn)
			continue
		}
		switch op.String() {
		case "create":
			{
				go FuntionCreateHandler(message, conn)
			}
		case "delete":
			{
				go FunctionDeleteHandler(message, conn)
			}
		case "update":
			{
				go FunctionUpdateHandler(message, conn)
			}
		}
	}
}


func FuntionCreateHandler(message []byte, conn *websocket.Conn) {
	function := &apiobject.Function{}
	function.UnMarshalJSON(message)
	log.Info("[FuntionCreateHandler] function: ", function)

	// check the parameters
	if function.Name == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("create: " + "function name is empty"))
		return
	}

	if function.Path == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("create: " + "function path is empty"))
	}

	err := activator.InitFunc(function.Name, function.Path)
	log.Info("[FuntionCreateHandler] init function finished")
	if err != nil {
		log.Error("[FunctionCreateHandler] error: ", err.Error())
		conn.WriteMessage(websocket.TextMessage, []byte("create: " +  err.Error()))
	} else {
		log.Info("[FunctionCreateHandler] success")
		conn.WriteMessage(websocket.TextMessage, []byte("create: " + "function create success"))
	}
}

// FunctionTriggerHandler the trigger format: {"name": "function name", "params": "function params"}
func FunctionTriggerHandler(message []byte, conn *websocket.Conn) {
	nameField := gjson.Get(string(message), "name")
	if !nameField.Exists() {
		conn.WriteMessage(websocket.TextMessage, []byte("execute: " + "function name is empty"))
		return
	}

	name := nameField.String()
	paramsField := gjson.Get(string(message), "params")
	if !paramsField.Exists() {
		conn.WriteMessage(websocket.TextMessage, []byte("execute: " + "function params is empty"))
		return
	}

	params := paramsField.String()
	log.Info("[FunctionTriggerHandler] name: ", name, ", params: ", params)
	result, err := activator.TriggerFunc(name, []byte(params))
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("execute: " + err.Error()))
		return
	}

	conn.WriteMessage(websocket.TextMessage, []byte("execute: " + string(result)))
}

func FunctionDeleteHandler(message []byte, conn *websocket.Conn) {
	function := &apiobject.Function{}
	function.UnMarshalJSON(message)
	log.Info("[FunctionDeleteHandler] function: ", function)

	// check the parameters
	if function.Name == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("delete: " + "function name is empty"))
		return
	}

	err := activator.DeleteFunc(function.Name)
	log.Info("[FunctionDeleteHandler] delete function finished")
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("delete: " + err.Error()))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("delete: " + "function delete success"))
	}

}

func FunctionUpdateHandler(message []byte, conn *websocket.Conn) {
	function := &apiobject.Function{}
	function.UnMarshalJSON(message)
	log.Info("[FunctionUpdateHandler] function: ", function)

	// delete the old function and create the new function
	err := activator.DeleteFunc(function.Name)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("update: " + err.Error()))
		return
	}
	log.Info("[FunctionUpdateHandler] delete function finished")
	err = activator.InitFunc(function.Name, function.Path)
	log.Info("[FunctionUpdateHandler] update function finished")
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("update: " + err.Error()))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("update: function update success"))
	}

}