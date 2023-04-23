package apimachinery

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/kubeapiserver/handlers"
	"minik8s/pkg/kubeapiserver/storage"
	"minik8s/pkg/kubeapiserver/watch"
	"strings"
)

type APIServer struct {
	HttpServer  *gin.Engine
	EtcdStorage *storage.EtcdStorage
}

func NewAPI() *APIServer {
	storage := storage.NewEtcdStorageNoParam()
	if storage == nil {
		return nil
	}

	return &APIServer{HttpServer: gin.Default(), EtcdStorage: storage}
}

// UpgradeToWebSocket the route middleware for update http request to websocket request
func (a *APIServer) UpgradeToWebSocket() gin.HandlerFunc {
	return func(c *gin.Context) {
		upgradeHeader := c.GetHeader("Upgrade")
		connectionHeader := c.GetHeader("Connection")
		if (strings.ToLower(upgradeHeader) == "websocket" && strings.Contains(strings.ToLower(connectionHeader), "upgrade")) || c.Query("watch") == "true" {
			// Stop the request processing
			c.Abort()

			// Get the key to watch
			// TODO: the attributes for watch may need to be supplied
			fullPath := c.Request.RequestURI
			// the resources that can be watched: pod, node
			var watchKey = ""
			if strings.Contains(fullPath, "pod") {
				// Check whether specify the namespace
				// the pod storage format
				namespace := c.Param("namespace")
				podName := c.Param("pod")

				if podName != "" && namespace != "" {
					watchKey = "/registry/pods/" + namespace + "/" + podName
				} else {
					watchKey = "/registry/pods/" + namespace
				}
			}
			// Setup a new websocket connection
			newWatcher, err := watch.NewWatchServer(c)
			if err != nil {
				log.Error("[UpgradeToWebSocket] fail to establish a new websocket connection, err: ", err)
				return
			}

			// add the watch server to the watch server map
			watchServerKey := c.ClientIP() + c.Request.RequestURI
			watch.WatchTable[watchServerKey] = newWatcher

			newWatcher.Watch(watchKey)
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

//func (a *APIServer) RegisterHandler(method string, path string, handler gin.HandlerFunc) {
//	// use middleware to upgrade http request to websocket request
//	a.HttpServer.Use(a.UpgradeToWebSocket())
//	switch method {
//	case "GET":
//		a.HttpServer.GET(path, handler)
//	case "POST":
//		a.HttpServer.POST(path, handler)
//	case "PUT":
//		a.HttpServer.PUT(path, handler)
//	case "DELETE":
//		a.HttpServer.DELETE(path, handler)
//	default:
//		panic("invalid HTTP method")
//	}
//}

func (a *APIServer) Run(addr string) error {
	for _, route := range handlers.HandlerTable {
		a.RegisterHandler(route)
	}
	return a.HttpServer.Run(addr)
}
