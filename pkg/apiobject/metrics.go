package apiobject

import (
	"encoding/json"
	"minik8s/pkg/apiobject/utils"
)

// NodeMetrics sets resource usage metrics of a node.
type NodeMetrics struct {
	APIVersion string   `json:"apiVersion,omitempty"`
	Data       MetaData `json:"metadata"`

	// The following fields define time interval from which metrics were
	// collected from the interval [Timestamp-Window, Timestamp].
	Timestamp utils.Time     `json:"timestamp"`
	Window    utils.Duration `json:"window"`

	// The memory usage is the memory working set.
	Usage ResourceList `json:"resources"`
}

/*
{
  "kind": "PodMetrics",
  "apiVersion": "metrics.k8s.io/v1beta1",
  "metadata": {
    "name": "svddb-servixxxxxxxxxxxxxxx4b4-bzs84",
    "namespace": "vddb",
    "selfLink": "/apis/metrics.k8s.io/v1beta1/namespaces/vddb/pods/svddbxxxxxxxxxxxxxxxx-bzs84",
    "creationTimestamp": "2022-12-16T14:40:46Z"
  },
  "timestamp": "2022-12-16T14:40:09Z",
  "window": "30s",
  "containers": [
    {
      "name": "svdxxxxxxxxrators",
      "usage": { "cpu": "2575748239n", "memory": "1257180Ki" }
    }
  ]
}
*/

type PodMetrics struct {
	APIVersion string   `json:"apiVersion,omitempty"`
	Data       MetaData `json:"metadata"`

	// The following fields define time interval from which metrics were
	// collected from the interval [Timestamp-Window, Timestamp].
	Timestamp utils.Time     `json:"timestamp"`
	Window    utils.Duration `json:"window"`

	// Metrics for all containers are collected within the same time window.
	Containers []ContainerMetrics `json:"containers"`
}

type ContainerMetrics struct {
	// Container name corresponding to the one from pod.spec.containers.
	Name string `json:"name"`
	// The memory usage is the memory working set.
	Usage ResourceList `json:"resources"`
}

type ResourceList map[ResourceName]utils.Quantity

func (p *PodMetrics) MarshalJSON() ([]byte, error) {
	type Alias PodMetrics
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	})
}

func (p *PodMetrics) UnMarshalJSON(data []byte) error {
	type Alias PodMetrics
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

func (n *NodeMetrics) MarshalJSON() ([]byte, error) {
	type Alias NodeMetrics
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(n),
	})
}

func (n *NodeMetrics) UnMarshalJSON(data []byte) error {
	type Alias NodeMetrics
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(n),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}
