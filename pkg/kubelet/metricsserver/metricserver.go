package metricsserver

import (
	"github.com/gin-gonic/gin"
)

type MetricsServer struct {
	HttpServer *gin.Engine
}

func NewMetricsServer() *MetricsServer {
	return &MetricsServer{HttpServer: gin.Default()}
}

func (s *MetricsServer) RegisterHandler(r route) {
	switch r.Method {
	case "GET":
		s.HttpServer.GET(r.Path, r.Handler)
	case "POST":
		s.HttpServer.POST(r.Path, r.Handler)
	case "PUT":
		s.HttpServer.PUT(r.Path, r.Handler)
	case "DELETE":
		s.HttpServer.DELETE(r.Path, r.Handler)
	}
}

func (s *MetricsServer) Run(addr string) error {
	for _, r := range handlerTable {
		s.RegisterHandler(r)
	}
	return s.HttpServer.Run(addr)
}
