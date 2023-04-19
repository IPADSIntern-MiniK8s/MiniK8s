package handlers

var HandlerTable = map[string]Route{
	"POST ^/api/v1/namespaces/([^/]+)/pods$":           {Path: "/api/v1/namespaces/:namespace/pods", Method: "POST", Handler: CreatePodHandler},                   // POST, create a pod
	"GET ^/api/v1/namespaces/([^/]+)/pods/([^/]+)$":    {Path: "/api/v1/namespaces/:namespace/pods/:name", Method: "GET", Handler: GetPodHandler},                 // GET, get a pod
	"DELETE ^/api/v1/namespaces/([^/]+)/pods/([^/]+)$": {Path: "/api/v1/namespaces/:namespace/pods/:name", Method: "DELETE", Handler: DeletePodHandler},           // DELETE, delete a pod                                                          // DELETE, delete a pod
	"PUT ^/api/v1/namespaces/([^/]+)/pods/status$":     {Path: "/api/v1/namespaces/:namespace/pods/:name/status", Method: "PUT", Handler: UpdatePodStatusHandler}, // PUT, update a pod status after a pod is created

	"POST ^/api/v1/nodes":         {Path: "/api/v1/nodes", Method: "POST", Handler: RegisterNodeHandler},       // POST, register a node
	"GET ^/api/v1/nodes":          {Path: "/api/v1/nodes", Method: "GET", Handler: GetNodesHandler},            // GET, list all nodes
	"GET ^/api/v1/nodes/([^/]+)$": {Path: "/api/v1/nodes/:name", Method: "GET", Handler: GetNodeByNameHandler}, // GET, get a node
}
