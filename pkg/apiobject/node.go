package apiobject

import (
	"encoding/json"
)

// Node
// a node struct for k8s node watch file
/*
{
  "kind": "Node",
  "apiVersion": "v1",
  "metadata": {
    "name": "node-1",
    "selfLink": "/api/v1/nodes/node-1",
    "uid": "f332a12e-3b38-11ec-abea-42010a800002",
    "resourceVersion": "123456",
    "creationTimestamp": "2022-10-24T10:00:00Z",
    "labels": {
      "key1": "value1",
      "key2": "value2"
    },
    "annotations": {
      "description": "This is a node for testing purposes"
    }
  },
  "spec": {
    "podCIDR": "10.244.0.0/24",
    "podCIDRs": [
      "10.244.0.0/24"
    ],
    "providerID": "aws://us-east-1/i-0123456789abcdef",
    "unschedulable": false,
    "taints": [
      {
        "key": "node-role",
        "value": "worker",
        "effect": "NoSchedule"
      }
    ]
  },
  "status": {
    "capacity": {
      "cpu": "4",
      "memory": "16Gi"
    },
    "allocatable": {
      "cpu": "4",
      "memory": "15Gi",
      "pods": "110"
    },
    "conditions": [
      {
        "type": "Ready",
        "status": "True",
        "lastHeartbeatTime": "2022-10-24T10:05:00Z",
        "lastTransitionTime": "2022-10-24T10:00:00Z",
        "reason": "KubeletReady",
        "message": "kubelet is ready"
      }
    ],
    "addresses": [
      {
        "type": "InternalIP",
        "address": "10.0.0.1"
      },
      {
        "type": "Hostname",
        "address": "node-1"
      }
    ],
    "daemonEndpoints": {
      "kubeletEndpoint": {
        "Port": 10250
      }
    },
    "nodeInfo": {
      "architecture": "amd64",
      "operatingSystem": "linux",
      "osImage": "Ubuntu 20.04.3 LTS",
      "kernelVersion": "5.11.0-38-generic",
      "kubeletVersion": "v1.22.2",
      "containerRuntimeVersion": "docker://20.10.8",
      "kubeProxyVersion": "v1.22.2"
    },
    "images": [
      {
        "names": [
          "nginx:1.21.4"
        ],
        "sizeBytes": 58590473
      },
      {
        "names": [
          "busybox:1.34.1"
        ],
        "sizeBytes": 1297162
      }
    ]
  }
}

*/

type Node struct {
	APIVersion string     `json:"apiVersion,omitempty"`
	Kind       string     `json:"kind,omitempty"`
	Data       MetaData   `json:"metadata,omitempty"`
	Spec       NodeSpec   `json:"spec,omitempty"`
	Status     NodeStatus `json:"status,omitempty"`
}

type NodeSpec struct {
	PodCIDR       string   `json:"podCIDR,omitempty"`
	PodCIDRs      []string `json:"podCIDRs,omitempty"`
	Unschedulable bool     `json:"unschedulable,omitempty"`
	Taints        []Taint  `json:"taints,omitempty"`
	ProviderID    string   `json:"providerID,omitempty"`
	//NodeSelector  map[string]string `json:"nodeSelector,omitempty"`
	//PodResources  PodResources      `json:"podResources,omitempty"`
}

type Taint struct {
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect,omitempty"`
}

type PodResources struct {
	Name   string `json:"name,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

type NodeStatus struct {
	Capability  map[string]string `json:"capacity,omitempty"`
	Allocatable map[string]string `json:"allocatable,omitempty"` // can be used to calculate the available resources, for scheduling
	Conditions  []Condition       `json:"conditions,omitempty"`
	Addresses   []Address         `json:"addresses,omitempty"`
	//DaemonEnd   DaemonEnd         `json:"daemonEndpoints,omitempty"`
	//NodeInfo    NodeInfo          `json:"nodeInfo,omitempty"`
	Images []Image `json:"images,omitempty"`
}

type Condition struct {
	Status             string `json:"status,omitempty"`
	LastHeartbeatTime  string `json:"lastHeartbeatTime,omitempty"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	Reason             string `json:"reason,omitempty"`
	Message            string `json:"message,omitempty"`
}

type NodeInfo struct {
	Architecture            string `json:"architecture,omitempty"`
	OperatingSystem         string `json:"operatingSystem,omitempty"`
	OsImage                 string `json:"osImage,omitempty"`
	KernelVersion           string `json:"kernelVersion,omitempty"`
	KubeletVersion          string `json:"kubeletVersion,omitempty"`
	ContainerRuntimeVersion string `json:"containerRuntimeVersion,omitempty"`
	KubeProxyVersion        string `json:"kubeProxyVersion,omitempty"`
}

type Address struct {
	Type    string `json:"type,omitempty"`
	Address string `json:"address,omitempty"`
}

type DaemonEnd struct {
	KubeletEndpoint KubeletEndpoint `json:"kubeletEndpoint,omitempty"`
}

type KubeletEndpoint struct {
	Port int `json:"port,omitempty"`
}

type Image struct {
	Names     []string `json:"names,omitempty"`
	SizeBytes int      `json:"sizeBytes,omitempty"`
}

func (n *Node) GetNode() *Node {
	return n
}

func (n *Node) MarshalJSON() ([]byte, error) {
	type Alias Node
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(n),
	})
}

// UnMarshalJSON json and store in current object
func (n *Node) UnMarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &n)
	if err != nil {
		return err
	}
	return nil
}
