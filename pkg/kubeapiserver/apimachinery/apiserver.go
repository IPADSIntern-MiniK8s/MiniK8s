package apimachinery

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/kubeapiserver/storage"
)

type APIServer struct {
	HttpServer *gin.Engine
	Watchers   map[string]*WatchServer // key: identifier for node/controller and etc.  value: websocket connection
	Storage    *storage.EtcdStorage
}

func NewAPI() *APIServer {
	storage := storage.NewEtcdStorageNoParam()
	if storage == nil {
		return nil
	}

	return &APIServer{HttpServer: gin.Default(), Storage: storage}
}

func (a *APIServer) RegisterHandler(method string, path string, handler gin.HandlerFunc) {
	// first filter watch request

	switch method {
	case "GET":
		a.HttpServer.GET(path, handler)
	case "POST":
		a.HttpServer.POST(path, handler)
	case "PUT":
		a.HttpServer.PUT(path, handler)
	case "DELETE":
		a.HttpServer.DELETE(path, handler)
	default:
		panic("invalid HTTP method")
	}
}

func (a *APIServer) Run(addr string) error {
	return a.HttpServer.Run(addr)
}
