package utils

import (
	"fmt"
	"github.com/coreos/etcd/Godeps/_workspace/src/github.com/golang/glog"
	"github.com/gorilla/websocket"
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
				syncFunc.HandleCreate(message)
			}
		case "DELETE":
			{
				syncFunc.HandleDelete(message)
			}
		case "UPDATE":
			{
				syncFunc.HandleUpdate(message)
			}
		}

	}
}

func CreateObject(obj apiobject.Object, ty ObjType, ns string) {
	res, _ := obj.MarshalJSON()
	//POST /api/v1/namespaces/{namespace}/{resource}"
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s", ApiServerIp, ns, ty)
	if info, err := SendRequest("POST", res, url); err != nil {
		glog.Error("create object ", info)
	}
}

func UpdateObject(obj apiobject.Object, ty ObjType, ns string, name string) {
	res, _ := obj.MarshalJSON()
	//POST /api/v1/namespaces/{namespace}/{resource}/{name}/update"
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s/%s/update", ApiServerIp, ns, ty, name)
	if info, err := SendRequest("POST", res, url); err != nil {
		glog.Error("create object ", info)
	}
}

func DeleteObject(ty ObjType, ns string, name string) {
	//DELETE /api/v1/namespaces/{namespace}/{resource}"
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s/%s", ApiServerIp, ns, ty, name)
	if info, err := SendRequest("DELETE", nil, url); err != nil {
		glog.Error("delete object ", info)
	}
}

func GetObjects(ty ObjType) string {
	//GET /api/v1/pods
	url := fmt.Sprintf("http://%s/api/v1/%s", ApiServerIp, ty)
	var str []byte
	if info, err := SendRequest("GET", str, url); err != nil {
		glog.Error("get object ", info)
		return info
	} else {
		return info
	}
}
