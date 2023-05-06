package apiobject

// Service
// a sevice struct for k8s service config file
/*
apiVersion: v1
kind: Service
metadata:
  name: service-practice
spec:
  selector:
    app: deploy-practice
  type: ClusterIP
  ports:
  - name: service-port1
    protocol: TCP
    port: 8080 # 对外暴露的端口
    targetPort: p1 # 转发的端口，pod对应的端口
  - name: service-port2
    protocol: TCP
    port: 3000 # 对外暴露的端口
    targetPort: p2 # 转发的端口，pod对应的端口
*/

// Service is a named abstraction of software service (for example, mysql) consisting of local port
// (for example 3306) that the proxy listens on, and the selector that determines which pods
// will answer requests sent through the proxy.
type Service struct {
	APIVersion string `json:"apiVersion,omitempty"`

	Data MetaData `json:"metadata"`

	// Spec defines the behavior of a service.
	// +optional
	Spec ServiceSpec `json:"spec,omitempty"`

	// Status represents the current status of a service.
	// +optional
	//Status ServiceStatus `json:"apiVersion,omitempty"`
}

// ServiceSpec describes the attributes that a user creates on a service
type ServiceSpec struct {
	// Type determines how the Service is exposed. Only ClusterIP is valid now.
	// "ClusterIP" allocates a cluster-internal IP address for load-balancing to
	// endpoints. Endpoints are determined by the selector or if that is not
	// specified, by manual construction of an Endpoints object. If clusterIP is
	// "None", no virtual IP is allocated and the endpoints are published as a
	// set of endpoints rather than a stable IP.
	Type ServiceType `json:"type,omitempty"`

	// Required: The list of ports that are exposed by this service.
	Ports []ServicePort `json:"ports"`

	// Route service traffic to pods with label keys and values matching this
	// selector. If empty or not present, the service is assumed to have an
	// external process managing its endpoints, which Kubernetes will not
	// modify. Only applies to types ClusterIP, NodePort, and LoadBalancer.
	// Ignored if type is ExternalName.
	Selector map[string]string `json:"selector"`

	// ClusterIP is the IP address of the service and is usually assigned
	// randomly by the master. If an address is specified manually and is not in
	// use by others, it will be allocated to the service; otherwise, creation
	// of the service will fail. This field can not be changed through updates.
	// Valid values are "None", empty string (""), or a valid IP address. "None"
	// can be specified for headless services when proxying is not required.
	// Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if
	// type is ExternalName.
	// +optional
	ClusterIP string `json:"clusterIP"`
}

// ServicePort represents the port on which the service is exposed
type ServicePort struct {
	// Optional if only one ServicePort is defined on this service: The
	// name of this port within the service.  This must be a DNS_LABEL.
	// All ports within a ServiceSpec must have unique names.  This maps to
	// the 'Name' field in EndpointPort objects.
	Name string `json:"name"`

	// The IP protocol for this port.  Supports "TCP", "UDP", and "SCTP".
	Protocol Protocol `json:"protocol"`

	// The port that will be exposed on the service.
	Port int32 `json:"port"`

	// Optional: The target port on pods selected by this service.  If this
	// is a string, it will be looked up as a named port in the target
	// Pod's container ports.  If this is not specified, the value
	// of the 'port' field is used (an identity map).
	// This field is ignored for services with clusterIP=None, and should be
	// omitted or set equal to the 'port' field.
	TargetPort IntOrString `json:"targetPort,omitempty"`
}

type ServiceType string

const (
	// ServiceTypeClusterIP means a service will only be accessible inside the
	// cluster, via the ClusterIP.
	ServiceTypeClusterIP ServiceType = "ClusterIP"

	// ServiceTypeNodePort means a service will be exposed on one port of
	// every node, in addition to 'ClusterIP' type.
	ServiceTypeNodePort ServiceType = "NodePort"

	// ServiceTypeLoadBalancer means a service will be exposed via an
	// external load balancer (if the cloud provider supports it), in addition
	// to 'NodePort' type.
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"

	// ServiceTypeExternalName means a service consists of only a reference to
	// an external name that kubedns or equivalent will return as a CNAME
	// record, with no exposing or proxying of any pods involved.
	ServiceTypeExternalName ServiceType = "ExternalName"
)

type Protocol string

const (
	// ProtocolTCP is the TCP protocol.
	ProtocolTCP Protocol = "TCP"
	// ProtocolUDP is the UDP protocol.
	ProtocolUDP Protocol = "UDP"
	// ProtocolSCTP is the SCTP protocol.
	ProtocolSCTP Protocol = "SCTP"
)

type IntOrString interface{}

//type ServiceStatus struct {
//	// Current service condition
//	// +optional
//	Conditions []metav1.Condition
//}

//type Condition struct {
//	// type of condition in CamelCase or in foo.example.com/CamelCase.
//	// ---
//	// Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be
//	// useful (see .node.status.conditions), the ability to deconflict is important.
//	// The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)
//	// +required
//	// +kubebuilder:validation:Required
//	// +kubebuilder:validation:Pattern=`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$`
//	// +kubebuilder:validation:MaxLength=316
//	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
//	// status of the condition, one of True, False, Unknown.
//	// +required
//	// +kubebuilder:validation:Required
//	// +kubebuilder:validation:Enum=True;False;Unknown
//	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status"`
//	// observedGeneration represents the .metadata.generation that the condition was set based upon.
//	// For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
//	// with respect to the current state of the instance.
//	// +optional
//	// +kubebuilder:validation:Minimum=0
//	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,3,opt,name=observedGeneration"`
//	// lastTransitionTime is the last time the condition transitioned from one status to another.
//	// This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.
//	// +required
//	// +kubebuilder:validation:Required
//	// +kubebuilder:validation:Type=string
//	// +kubebuilder:validation:Format=date-time
//	LastTransitionTime Time `json:"lastTransitionTime" protobuf:"bytes,4,opt,name=lastTransitionTime"`
//	// reason contains a programmatic identifier indicating the reason for the condition's last transition.
//	// Producers of specific condition types may define expected values and meanings for this field,
//	// and whether the values are considered a guaranteed API.
//	// The value should be a CamelCase string.
//	// This field may not be empty.
//	// +required
//	// +kubebuilder:validation:Required
//	// +kubebuilder:validation:MaxLength=1024
//	// +kubebuilder:validation:MinLength=1
//	// +kubebuilder:validation:Pattern=`^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$`
//	Reason string `json:"reason" protobuf:"bytes,5,opt,name=reason"`
//	// message is a human readable message indicating details about the transition.
//	// This may be an empty string.
//	// +required
//	// +kubebuilder:validation:Required
//	// +kubebuilder:validation:MaxLength=32768
//	Message string `json:"message" protobuf:"bytes,6,opt,name=message"`
//}
