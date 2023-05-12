package apimachinery

import "github.com/gin-gonic/gin"

// ServiceRoutes the route information for service
type ServiceRoutes struct {
	ServiceName string
	Routes      []Route
}

// Route specific route
type Route struct {
	Path    string
	Method  string
	Handler gin.HandlerFunc
}

func (r *Route) register(engine *gin.Engine) {
	switch r.Method {
	case "GET":
		engine.GET(r.Path, r.Handler)
	case "POST":
		engine.POST(r.Path, r.Handler)
	case "PUT":
		engine.PUT(r.Path, r.Handler)
	case "DELETE":
		engine.DELETE(r.Path, r.Handler)
	default:
		panic("invalid HTTP method")
	}

}

func (serv *ServiceRoutes) registerRoutes(engine *gin.Engine) {
	for i := range serv.Routes {
		serv.Routes[i].register(engine)
	}
}

// WatchFilter route filter for "watch=true" query parameter
func WatchFilter(c *gin.Context) bool {
	if c.Query("watch") == "true" {
		return true
	} else {
		return false
	}
}
