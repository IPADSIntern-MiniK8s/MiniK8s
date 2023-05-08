package apiobject

import "encoding/json"

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
	Status ServiceStatus `json:"status,omitempty"`
}

// ServiceSpec describes the attributes that a user creates on a service
type ServiceSpec struct {
	// Type determines how the Service is exposed. Only ClusterIP is valid now.
	Type ServiceType `json:"type,omitempty"`

	// Required: The list of ports that are exposed by this service.
	Ports []ServicePort `json:"ports"`

	// Route service traffic to pods with label keys and values matching this selector.
	Selector map[string]string `json:"selector"`

	// ClusterIP is the IP address of the service and is usually assigned randomly by the master.
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
	TargetPort string `json:"targetPort"`
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

type ServiceStatus struct {
	/*
		CREATING: 等待分配cluster ip
		CREATED: cluster ip分配完成
	*/
	Phase string `json:"phase,omitempty"`
}

func (s *Service) UnMarshalJSON(data []byte) error {
	type Alias Service
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

func (s *Service) MarshalJSON() ([]byte, error) {
	type Alias Service
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
}
