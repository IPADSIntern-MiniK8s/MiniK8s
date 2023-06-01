package utils

import (
	"fmt"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type SyncFunc interface {
	GetType() config.ObjType
	HandleCreate(message []byte)
	HandleDelete(message []byte)
	HandleUpdate(message []byte)
}

func Sync(syncFunc SyncFunc) {

	url := fmt.Sprintf("ws://%s/api/v1/watch/%s", config.ApiServerIp, syncFunc.GetType())
	for {
		err := connect(url, syncFunc)
		if err != nil {
			fmt.Println("WebSocket连接错误:", err)
		}

		fmt.Println("连接中断，等待重新连接...")
		time.Sleep(5 * time.Second) // 等待5秒后进行重连
	}

}

func connect(url string, syncFunc SyncFunc) error {
	// 建立WebSocket连接
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("WebSocket连接失败：", err)
		return err
	} else {
		fmt.Println("WebSocket连接成功")
	}

	defer conn.Close()

	// 不断地接收消息并处理
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("读取消息失败：", err)
			return err
		}
		if len(message) == 0 {
			continue
		}
		// fmt.Printf("[client %s] %s\n", syncFunc.GetType(), message)

		op := gjson.Get(string(message), "metadata.resourcesVersion")
		switch op.String() {
		case "create":
			{
				go syncFunc.HandleCreate(message)
			}
		case "delete":
			{
				go syncFunc.HandleDelete(message)
			}
		case "update":
			{
				go syncFunc.HandleUpdate(message)
			}
		}

	}
}

func CreateObject(obj apiobject.Object, ty config.ObjType, ns string) {
	if ns == "" {
		ns = "default"
	}
	res, _ := obj.MarshalJSON()
	// log.Info("[create obj]", string(res))
	//POST /api/v1/namespaces/{namespace}/{resource}"
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s", config.ApiServerIp, ns, ty)
	if info, err := SendRequest("POST", res, url); err != nil {
		log.Error("create object ", info)
	}
}

func UpdateObject(obj apiobject.Object, ty config.ObjType, ns string, name string) {
	if ns == "" {
		ns = "default"
	}
	res, _ := obj.MarshalJSON()
	// log.Info("[update obj]", string(res))
	//POST /api/v1/namespaces/{namespace}/{resource}/{name}/update"
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s/%s/update", config.ApiServerIp, ns, ty, name)
	if info, err := SendRequest("POST", res, url); err != nil {
		log.Error("update object ", info)
	}
}

func UpdateObjectStatus(obj apiobject.Object, ty config.ObjType, ns string, name string) {
	if ns == "" {
		ns = "default"
	}
	res, _ := obj.MarshalJSON()
	// log.Info("[update obj status]", string(res))
	//POST /api/v1/namespaces/{namespace}/{resource}/{name}/status"
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s/%s/status", config.ApiServerIp, ns, ty, name)
	if info, err := SendRequest("POST", res, url); err != nil {
		log.Error("update object ", info)
	}
}

func DeleteObject(ty config.ObjType, ns string, name string) {
	if ns == "" {
		ns = "default"
	}
	// log.Info("[delete obj]", name)
	//DELETE /api/v1/namespaces/{namespace}/{resource}"
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s/%s", config.ApiServerIp, ns, ty, name)
	if info, err := SendRequest("DELETE", nil, url); err != nil {
		log.Error("delete object ", info)
	}
}

func GetObject(ty config.ObjType, ns string, name string) string {
	if ns == "" {
		ns = "default"
	}
	// log.Info("[get obj]", name)
	//GET /api/v1/pods
	var url string
	if ns != "nil" {
		if name == "" {
			url = fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s", config.ApiServerIp, ns, ty)
		} else {
			url = fmt.Sprintf("http://%s/api/v1/namespaces/%s/%s/%s", config.ApiServerIp, ns, ty, name)
		}
	} else {
		if name == "" {
			url = fmt.Sprintf("http://%s/api/v1/%s", config.ApiServerIp, ty)
		} else {
			url = fmt.Sprintf("http://%s/api/v1/%s/%s", config.ApiServerIp, ty, name)
		}
	}

	var str []byte
	if info, err := SendRequest("GET", str, url); err != nil {
		log.Error("[get obj]", info)
		return ""
	} else {
		return info
	}
}


