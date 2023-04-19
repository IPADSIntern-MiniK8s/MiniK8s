package apiobject

import (
	"encoding/json"
)

// Node
// a node struct for k8s node config file
/*
apiVersion: v1
kind: Node
metadata:
  name: my-node
  labels:
    env: production
spec:
  podCIDR: 10.244.0.0/24
  podCIDRs:
  - 10.244.0.0/24
  unschedulable: false
  taints:
  - key: dedicated
    value: my-node
    effect: NoSchedule
  - key: special
    value: node
    effect: PreferNoSchedule
  configSource:
    configMap:
      name: kubelet-config
      namespace: kube-system
      kubeletConfigKey: config.yaml
  providerID: my-cloud-provider://i-0123456789abcdef
  nodeSelector:
    size: Large
  topologySpreadConstraints:
  - maxSkew: 1
    topologyKey: kubernetes.io/hostname
    whenUnsatisfiable: ScheduleAnyway
    labelSelector:
      matchExpressions:
      - key: environment
        operator: In
        values:
        - production
  podResources:
    - name: cpu
      weight: 2
    - name: memory
      weight: 1
*/
type Node struct {
	APIVersion string   `json:"apiVersion,omitempty"`
	Kind       string   `json:"kind,omitempty"`
	Data       MetaData `json:"metadata,omitempty"`
	Spec       NodeSpec `json:"spec,omitempty"`
}

type NodeSpec struct {
	PodCIDR       string            `json:"podCIDR,omitempty"`
	PodCIDRs      []string          `json:"podCIDRs,omitempty"`
	Unschedulable bool              `json:"unschedulable,omitempty"`
	Taints        []Taint           `json:"taints,omitempty"`
	ProviderID    string            `json:"providerID,omitempty"`
	NodeSelector  map[string]string `json:"nodeSelector,omitempty"`
	PodResources  PodResources      `json:"podResources,omitempty"`
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

func (n *Node) UnMarshalJSON(data []byte) error {
	type Alias Node
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(n),
	}
	return json.Unmarshal(data, aux)
}
