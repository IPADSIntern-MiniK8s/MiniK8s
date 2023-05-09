package utils

import (
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"minik8s/pkg/apiobject"
)

type SyncFunc interface {
	GetType() ObjType
	HandleCreate(message []byte)
	HandleDelete(message []byte)
	HandleUpdate(message []byte)
}

func Sync(syncFunc SyncFunc) {
	// 建立WebSocket连接
	url := fmt.Sprintf("ws://%s/api/v1/watch/%s", ApiServerIp, syncFunc.GetType())
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("WebSocket连接失败：", err)
		return
	} else {
		fmt.Println("WebSocket连接成功")
	}

	// 不断地接收消息并处理
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("读取消息失败：", err)
			return
		}
		fmt.Printf("%s\n", message)

		op := gjson.Get("metadata.resourcesVersion", string(message))
		switch op.String() {
		case "CREATE":
			{
				go syncFunc.HandleCreate(message)
			}
		case "DELETE":
			{
				go syncFunc.HandleDelete(message)
			}
		case "UPDATE":
			{
				go syncFunc.HandleUpdate(message)
			}
		}

	}
}

func CreateObject(obj apiobject.Object, ty ObjType, ns string) {
	if ns == "" {
		ns = "default"
	}
	res, _ := obj.MarshalJSON()
	fmt.Println("[create obj]", string(res))
	//POST /api/v1/namespaces/{namespace}/{resource}"
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s", ApiServerIp, ns, ty)
	if info, err := SendRequest("POST", res, url); err != nil {
		log.Error("create object ", info)
	}
}

func UpdateObject(obj apiobject.Object, ty ObjType, ns string, name string) {
	if ns == "" {
		ns = "default"
	}
	res, _ := obj.MarshalJSON()
	//POST /api/v1/namespaces/{namespace}/{resource}/{name}/update"
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s/%s/update", ApiServerIp, ns, ty, name)
	if info, err := SendRequest("POST", res, url); err != nil {
		log.Error("create object ", info)
	}
}

func DeleteObject(ty ObjType, ns string, name string) {
	if ns == "" {
		ns = "default"
	}
	//DELETE /api/v1/namespaces/{namespace}/{resource}"
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s/%s", ApiServerIp, ns, ty, name)
	if info, err := SendRequest("DELETE", nil, url); err != nil {
		log.Error("delete object ", info)
	}
}

func GetObject(ty ObjType, ns string, name string) string {
	if ns == "" {
		ns = "default"
	}
	//GET /api/v1/pods
	var url string
	if name == "" {
		url = fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s", ApiServerIp, ns, ty)
	} else {
		url = fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s/%s", ApiServerIp, ns, ty, name)
	}
	var str []byte
	if info, err := SendRequest("GET", str, url); err != nil {
		log.Error("get object ", info)
		return info
	} else {
		return info
	}
}
