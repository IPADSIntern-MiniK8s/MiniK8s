package kubeapiserver

import (
	"github.com/gin-gonic/gin"
)

type APIServer struct {
	engine *gin.Engine
}

func NewAPI() *APIServer {
	return &APIServer{engine: gin.Default()}
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
