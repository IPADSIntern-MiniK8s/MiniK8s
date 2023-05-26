package handlers

var HandlerTable = [...]Route{
	{Path: "/api/v1/namespaces/:namespace/pods", Method: "POST", Handler: CreatePodHandler},                    // POST, create a pod
	{Path: "/api/v1/namespaces/:namespace/pods/:name", Method: "GET", Handler: GetPodHandler},                  // GET, get a pod
	{Path: "/api/v1/namespaces/:namespace/pods", Method: "GET", Handler: GetPodsHandler},                       // GET, list all pods
	{Path: "/api/v1/namespaces/:namespace/pods/:name", Method: "DELETE", Handler: DeletePodHandler},            // DELETE, delete a pod
	{Path: "/api/v1/namespaces/:namespace/pods/:name/update", Method: "POST", Handler: UpdatePodStatusHandler}, // POST, update a pod
	{Path: "/api/v1/pods", Method: "GET", Handler: GetAllPodsHandler},                                          // GET, get all pods

	{Path: "/api/v1/nodes", Method: "POST", Handler: RegisterNodeHandler},                         // POST, register a node
	{Path: "/api/v1/watch/:resource/:namespaces/:name", Method: "GET", Handler: NodeWatchHandler}, // GET, watch a resource
	{Path: "/api/v1/nodes", Method: "GET", Handler: GetNodesHandler},                              // GET, list all nodes
	{Path: "/api/v1/nodes/:name", Method: "GET", Handler: GetNodeByNameHandler},                   // GET, get a node

	{Path: "/api/v1/namespaces/:namespace/services", Method: "POST", Handler: CreateServiceHandler},             // POST, create a service
	{Path: "/api/v1/namespaces/:namespace/services/:name", Method: "GET", Handler: GetServiceHandler},           // GET, get a service
	{Path: "/api/v1/namespaces/:namespace/services", Method: "GET", Handler: GetServicesHandler},                // GET, list all services
	{Path: "/api/v1/namespaces/:namespace/services/:name", Method: "DELETE", Handler: DeleteServiceHandler},     // DELETE, delete a service
	{Path: "api/v1/namespaces/:namespace/services/:name/update", Method: "POST", Handler: UpdateServiceHandler}, // POST, update a service

	{Path: "/api/v1/namespaces/:namespace/endpoints", Method: "POST", Handler: CreateEndpointHandler},             // POST, create a endpoint
	{Path: "/api/v1/namespaces/:namespace/endpoints/:name", Method: "GET", Handler: GetEndpointHandler},           // GET, get a endpoint
	{Path: "/api/v1/namespaces/:namespace/endpoints", Method: "GET", Handler: GetEndpointsHandler},                // GET, list all endpoints in this namespace
	{Path: "/api/v1/namespaces/:namespace/endpoints/:name", Method: "DELETE", Handler: DeleteEndpointHandler},     // DELETE, delete a endpoint
	{Path: "api/v1/namespaces/:namespace/endpoints/:name/update", Method: "POST", Handler: UpdateEndpointHandler}, // POST, update a endpoint

	{Path: "/api/v1/dns", Method: "POST", Handler: CreateDNSRecordHandler},              // POST, create a dns
	{Path: "/api/v1/dns/:name", Method: "GET", Handler: GetDNSRecordHandler},            // GET, get a dns
	{Path: "/api/v1/dns", Method: "GET", Handler: GetDNSRecordsHandler},                 // GET, list all dns records
	{Path: "/api/v1/dns/:name", Method: "DELETE", Handler: DeleteDNSRecordHandler},      // DELETE, delete a dns record
	{Path: "/api/v1/dns/:name/update", Method: "POST", Handler: UpdateDNSRecordHandler}, // POST, update a dns record

	{Path: "/api/v1/namespaces/:namespace/replicas", Method: "POST", Handler: CreateReplicaHandler},              // POST, create a replica
	{Path: "/api/v1/namespaces/:namespace/replicas/:name", Method: "GET", Handler: GetReplicaHandler},            // GET, get a replica
	{Path: "/api/v1/namespaces/:namespace/replicas", Method: "GET", Handler: GetReplicasHandler},                 // GET, list all replicas
	{Path: "/api/v1/namespaces/:namespace/replicas/:name", Method: "DELETE", Handler: DeleteReplicaHandler},      // DELETE, delete a replica
	{Path: "/api/v1/namespaces/:namespace/replicas/:name/update", Method: "POST", Handler: UpdateReplicaHandler}, // POST, update a replica

	{Path: "/api/v1/functions", Method: "POST", Handler: UploadFunctionHandler},                // POST, create a function
	{Path: "/api/v1/functions/:name", Method: "GET", Handler: GetFunctionHandler},              // GET, get a function
	{Path: "/api/v1/functions/:name", Method: "DELETE", Handler: DeleteFunctionHandler},        // DELETE, delete a function
	{Path: "/api/v1/functions/:name/update", Method: "POST", Handler: UpdateFunctionHandler},   // POST, update a function
	{Path: "/api/v1/functions/:name/trigger", Method: "POST", Handler: TriggerFunctionHandler}, // POST, trigger a function
	{Path: "/api/v1/functions", Method: "GET", Handler: GetFunctionsHandler},                   // GET, list all functions

	{Path: "/api/v1/namespaces/:namespace/hpas", Method: "POST", Handler: CreateHpaHandler},                    // POST, create a hpa
	{Path: "/api/v1/namespaces/:namespace/hpas/:name", Method: "GET", Handler: GetHpaHandler},                  // GET, get a hpa
	{Path: "/api/v1/namespaces/:namespace/hpas", Method: "GET", Handler: GetHpasHandler},                       // GET, list all hpas
	{Path: "/api/v1/namespaces/:namespace/hpas/:name", Method: "DELETE", Handler: DeleteHpaHandler},            // DELETE, delete a hpa
	{Path: "/api/v1/namespaces/:namespace/hpas/:name/update", Method: "POST", Handler: UpdateHpaHandler},       // POST, update a hpa
	{Path: "/api/v1/hpas", Method: "GET", Handler: GetAllHpaHandler},                                           // GET, get all hpas
	{Path: "/api/v1/namespaces/:namespace/hpas/:name/status", Method: "POST", Handler: UpdateHpaStatusHandler}, // POST, update hpa status

	{Path: "/api/v1/workflows", Method: "POST", Handler: UploadWorkflowHandler},                // POST, create a workflow
	{Path: "/api/v1/workflows/:name", Method: "GET", Handler: GetWorkflowHandler},              // GET, get a workflow
	{Path: "/api/v1/workflows", Method: "GET", Handler: GetWorkflowsHandler},                   // GET, list all workflows
	{Path: "/api/v1/workflows/:name", Method: "DELETE", Handler: DeleteWorkflowHandler},        // DELETE, delete a workflow
	{Path: "/api/v1/workflows/:name/update", Method: "POST", Handler: UpdateWorkflowHandler},   // POST, update a workflow
	{Path: "/api/v1/workflows/:name/trigger", Method: "POST", Handler: TriggerWorkflowHandler}, // POST, trigger a workflow
}
