package apimachinery

import (
	"context"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/handlers"
	"minik8s/pkg/kubeapiserver/storage"
	"minik8s/pkg/kubeapiserver/watch"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type APIServer struct {
	HttpServer  *gin.Engine
	EtcdStorage *storage.EtcdStorage
}


func InitNodes(tool *storage.EtcdStorage) {
	// delete all nodes' info in etcd
	key := "/registry/nodes/"
	var nodes []apiobject.Node
	err := tool.GetList(context.Background(), key, &nodes)
	if err != nil {
		log.Info("[InitNodes] the node list is empty")
	} else {
		for _, node := range nodes {
			nodeKey := "/registry/nodes/" + node.Data.Name
			err := tool.Delete(context.Background(), nodeKey)
			if err != nil {
				log.Error("[InitNodes] delete node error: ", err)
			}
		}
	}

}

func NewAPI() *APIServer {
	storage := storage.NewEtcdStorageNoParam()
	if storage == nil {
		return nil
	}
	
	InitNodes(storage)
	return &APIServer{HttpServer: gin.Default(), EtcdStorage: storage}
}

// UpgradeToWebSocket the route middleware for update http request to websocket request
func (a *APIServer) UpgradeToWebSocket() gin.HandlerFunc {
	return func(c *gin.Context) {
		upgradeHeader := c.GetHeader("Upgrade")
		connectionHeader := c.GetHeader("Connection")
		sourceHeader := c.GetHeader("X-Source")
		if (strings.ToLower(upgradeHeader) == "websocket" && strings.Contains(strings.ToLower(connectionHeader), "upgrade")) || c.Query("watch") == "true" {
			// Stop the request processing
			c.Abort()

			// Get the key to watch
			// the resources that can be watched: pod, node, service...
			resource := c.Param("resource")
			namespace := c.Param("namespace")
			name := c.Param("name")
			var watchKey = "/registry/" + resource
			if namespace != "" {
				watchKey += "/" + namespace
				if name != "" {
					watchKey += "/" + name
				}
			}

			// set up a new websocket connection
			newWatcher, err := watch.NewWatchServer(c)
			if err != nil {
				log.Error("[UpgradeToWebSocket] fail to establish a new websocket connection, err: ", err)
				return
			}

			// add the watch server to the watch server map
			// only service and node watch need to add to the watch table, and all of them watch the all pods
			log.Info("[UpgradeToWebSocket] watchKey: ", watchKey)
			if sourceHeader != "" {
				watch.WatchTable[sourceHeader] = newWatcher
				log.Info("[NodeWatchHandler] watchTable size: ", len(watch.WatchTable))
			}

			// store the connection in the watchStorage
			list, ok := watch.WatchStorage.Load(watchKey)
			if ok {
				if threadList, ok := list.(*watch.ThreadSafeList); !ok {
					log.Error("[UpgradeToWebSocket] fail to convert the list to ThreadSafeList")
					return
				} else {
					threadList.PushBack(newWatcher)
					watch.WatchStorage.Store(watchKey, threadList)
				}

			} else {
				// create a list
				newList := watch.NewThreadSafeList()
				newList.PushBack(newWatcher)
				watch.WatchStorage.Store(watchKey, newList)
			}
		} else {
			// Continue with the request processing
			c.Next()
		}
	}
}

func (a *APIServer) RegisterHandler(route handlers.Route) {
	a.HttpServer.Use(a.UpgradeToWebSocket())
	switch route.Method {
	case "GET":
		a.HttpServer.GET(route.Path, route.Handler)
	case "POST":
		a.HttpServer.POST(route.Path, route.Handler)
	case "PUT":
		a.HttpServer.PUT(route.Path, route.Handler)
	case "DELETE":
		a.HttpServer.DELETE(route.Path, route.Handler)
	}
}


func (a *APIServer) Run(addr string) error {
	for _, route := range handlers.HandlerTable {
		a.RegisterHandler(route)
	}

	return a.HttpServer.Run(addr)
}
