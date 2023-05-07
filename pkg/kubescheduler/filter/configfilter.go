package filter

import (
	"minik8s/pkg/apiobject"
	"strconv"
)

// ConfigFilter is a filter that filter node by config information
type ConfigFilter struct {
	Name string
}

// PreFilter runs a set of functions against a pod. If any of the functions returns an error, the pod is rejected.
func (f *ConfigFilter) PreFilter(pod *apiobject.Pod) bool {
	// check whether the pod is empty
	if pod == nil || pod.Status.Phase == "" {
		return false
	}

	// check whether the pod is in pending state
	if pod.Status.Phase != apiobject.Pending {
		return false
	}

	return true
}

// checkNodeSelector checks whether the node has the label that the pod needs
func (f *ConfigFilter) checkNodeSelector(pod *apiobject.Pod, node *apiobject.Node) bool {
	// check whether the pod has node selector
	//if pod.Spec.NodeSelector == nil {
	//	return true
	//}

	// check whether the node has the label that the pod needs
	if node.Data.Labels == nil {
		return false
	}
	for key, value := range pod.Spec.NodeSelector {
		if node.Data.Labels[key] != value {
			return false
		}
	}
	return true
}

// getResourceRequest gets the resource that the pod needs
func (f *ConfigFilter) getResourceRequest(pod *apiobject.Pod) (float64, float64) {
	// check whether the pod has resource request
	if pod.Spec.Containers == nil || len(pod.Spec.Containers) == 0 ||
		(pod.Spec.Containers[0].Resources.Requests.Cpu == "" && pod.Spec.Containers[0].Resources.Requests.Memory == "") {
		return 0.0, 0.0
	}

	// calculate the resource that the pod needs
	totalCpu, totalMemory := 0.0, 0.0
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests.Cpu != "" {
			cpuRequest, _ := strconv.ParseFloat(container.Resources.Requests.Cpu, 64)
			totalCpu += cpuRequest
		}
		if container.Resources.Requests.Memory != "" {
			memoryRequest, _ := strconv.ParseFloat(container.Resources.Requests.Memory, 64)
			totalMemory += memoryRequest
		}
	}
	return totalCpu, totalMemory
}

func (f *ConfigFilter) checkResource(cpuRequest float64, memoryRequest float64, nodes []*apiobject.Node) []*apiobject.Node {
	// check whether the pod has resource request
	if cpuRequest == 0.0 && memoryRequest == 0.0 {
		return nodes
	}

	// check whether the node has enough resource
	result := make([]*apiobject.Node, 0)
	for _, node := range nodes {
		// if the
		if node.Data.Status.Allocatable.Cpu >= cpuRequest && node.Data.Status.Allocatable.Memory >= memoryRequest {
			result = append(result, node)
		}
	}
}

func (f *ConfigFilter) Filter(pod *apiobject.Pod, nodes []*apiobject.Node) []*apiobject.Node {
	//
}
