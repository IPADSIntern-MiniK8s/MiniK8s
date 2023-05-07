package handlers

var HandlerTable = [...]Route{
	{Path: "/api/v1/namespaces/:namespace/pods", Method: "POST", Handler: CreatePodHandler},                    // POST, create a pod
	{Path: "/api/v1/namespaces/:namespace/pods/:name", Method: "GET", Handler: GetPodHandler},                  // GET, get a pod
	{Path: "/api/v1/namespaces/:namespace/pods", Method: "GET", Handler: GetPodsHandler},                       // GET, list all pods
	{Path: "/api/v1/namespaces/:namespace/pods/:name", Method: "DELETE", Handler: DeletePodHandler},            // DELETE, delete a pod
	{Path: "/api/v1/namespaces/:namespace/pods/:name/update", Method: "POST", Handler: UpdatePodStatusHandler}, // POST, update a pod

	{Path: "/api/v1/nodes", Method: "POST", Handler: RegisterNodeHandler},                                    // POST, register a node
	{Path: "/api/v1/watch(/:source)?/:resource/:namespaces/:name", Method: "GET", Handler: NodeWatchHandler}, // GET, watch a resource
	{Path: "/api/v1/nodes", Method: "GET", Handler: GetNodesHandler},                                         // GET, list all nodes
	{Path: "/api/v1/nodes/:name", Method: "GET", Handler: GetNodeByNameHandler},                              // GET, get a node

	{Path: "/api/v1/namespaces/:namespace/services", Method: "POST", Handler: CreateServiceHandler},             // POST, create a service
	{Path: "/api/v1/namespaces/:namespace/services/:name", Method: "GET", Handler: GetServiceHandler},           // GET, get a service
	{Path: "/api/v1/namespaces/:namespace/services", Method: "GET", Handler: GetServicesHandler},                // GET, list all services
	{Path: "/api/v1/namespaces/:namespace/services/:name", Method: "DELETE", Handler: DeleteServiceHandler},     // DELETE, delete a service
	{Path: "api/v1/namespaces/:namespace/services/:name/update", Method: "POST", Handler: UpdateServiceHandler}, // POST, update a service
}
