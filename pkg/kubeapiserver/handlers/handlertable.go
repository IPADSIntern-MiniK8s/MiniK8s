package handlers

import "github.com/gin-gonic/gin"

var handlerTable = map[string]gin.HandlerFunc{
	"POST ^/api/v1/namespaces/([^/]+)/pods$":           CreatePodHandler, // POST, create a pod
	"GET ^/api/v1/namespaces/([^/]+)/pods/([^/]+)$":    GetPodHandler,    // GET, get a pod
	"DELETE ^/api/v1/namespaces/([^/]+)/pods/([^/]+)$": DeletePodHandler, // DELETE, delete a pod

}
