package kubeapiserver

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/gin-gonic/gin"
)

type APIServer struct {
	engine  *gin.Engine
	storage *clientv3.Client
}

func NewAPI() *APIServer {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://etcd:2379"},
	})
	if err != nil {
		return nil
	}

	return &APIServer{engine: gin.Default(), storage: client}
}

func (a *APIServer) RegisterHandler(method string, path string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		a.engine.GET(path, handler)
	case "POST":
		a.engine.POST(path, handler)
	case "PUT":
		a.engine.PUT(path, handler)
	case "DELETE":
		a.engine.DELETE(path, handler)
	default:
		panic("invalid HTTP method")
	}
}

func (a *APIServer) Run(addr string) error {
	return a.engine.Run(addr)
}
