package handlers

var HandlerTable = map[string]Route{
	"POST ^/api/v1/namespaces/([^/]+)/pods$":                {Path: "/api/v1/namespaces/:namespace/pods", Method: "POST", Handler: CreatePodHandler},                    // POST, create a pod
	"GET ^/api/v1/namespaces/([^/]+)/pods/([^/]+)$":         {Path: "/api/v1/namespaces/:namespace/pods/:name", Method: "GET", Handler: GetPodHandler},                  // GET, get a pod
	"GET ^/api/v1/namespaces/([^/]+)/pods$":                 {Path: "/api/v1/namespaces/:namespace/pods", Method: "GET", Handler: GetPodsHandler},                       // GET, list all pods
	"DELETE ^/api/v1/namespaces/([^/]+)/pods/([^/]+)$":      {Path: "/api/v1/namespaces/:namespace/pods/:name", Method: "DELETE", Handler: DeletePodHandler},            // DELETE, delete a pod                                                          // DELETE, delete a pod
	"POST ^/api/v1/namespaces/([^/]+)/pods/([^/]+)/update$": {Path: "/api/v1/namespaces/:namespace/pods/:name/update", Method: "POST", Handler: UpdatePodStatusHandler}, // POST, update a pod

	"POST ^/api/v1/nodes": {Path: "/api/v1/nodes", Method: "POST", Handler: RegisterNodeHandler}, // POST, register a node
	"GET ^/api/v1/watch/:source/:resource/*namespaces?/*name?": {Path: "/api/v1/watch(/:source)?/:resource/:namespaces/:name", Method: "GET", Handler: NodeWatchHandler}, // GET, watch a resource
	"GET ^/api/v1/nodes":          {Path: "/api/v1/nodes", Method: "GET", Handler: GetNodesHandler},            // GET, list all nodes
	"GET ^/api/v1/nodes/([^/]+)$": {Path: "/api/v1/nodes/:name", Method: "GET", Handler: GetNodeByNameHandler}, // GET, get a node
}
