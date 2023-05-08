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

// checkNodeStatus checks whether the node is ready and can be found
func (f *ConfigFilter) checkNodeStatus(nodes []*apiobject.Node) []*apiobject.Node {
	// check whether the node is ready and can be found
	result := make([]*apiobject.Node, 0)
	for _, node := range nodes {
		if node.Status.Conditions[0].Status == "Ready" {
			// check whether the node has the InternalIp field
			if node.Status.Addresses == nil || len(node.Status.Addresses) == 0 {
				continue
			}
			hasInternalIp := false
			for _, address := range node.Status.Addresses {
				if address.Type == "InternalIP" && address.Address != "" {
					hasInternalIp = true
					break
				}
			}
			if hasInternalIp {
				result = append(result, node)
			}
		}
	}
	return result
}

// checkNodeSelector checks whether the node has the label that the pod needs
func (f *ConfigFilter) checkNodeSelector(pod *apiobject.Pod, nodes []*apiobject.Node) []*apiobject.Node {
	// check whether the pod has node selector
	if pod.Spec.NodeSelector == nil {
		return nodes
	}

	// check whether the node has the label that the pod needs
	result := make([]*apiobject.Node, 0)
	for _, node := range nodes {
		if node.Data.Labels == nil {
			continue
		}
		for key, value := range pod.Spec.NodeSelector {
			if node.Data.Labels[key] != value {
				continue
			}
		}
		result = append(result, node)
	}

	return result
}

// GetResourceRequest gets the resource that the pod needs
func (f *ConfigFilter) GetResourceRequest(pod *apiobject.Pod) (float64, float64) {
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
		// if the according field is empty, it means that the node may has enough resource
		if node.Status.Allocatable == nil {
			result = append(result, node)
			continue
		}

		// check whether the node has enough CPU
		cpu, ok := node.Status.Allocatable["cpu"]
		if !ok {
			result = append(result, node)
			continue
		}
		cpuAvailable, _ := strconv.ParseFloat(cpu, 64)
		if cpuAvailable < cpuRequest {
			continue
		}

		// check whether the node has enough memory
		memory, ok := node.Status.Allocatable["memory"]
		if !ok {
			result = append(result, node)
			continue
		}
		memoryAvailable, _ := strconv.ParseFloat(memory, 64)
		if memoryAvailable < memoryRequest {
			continue
		}

		result = append(result, node)
	}

	return result
}

func (f *ConfigFilter) Filter(pod *apiobject.Pod, nodes []*apiobject.Node) []*apiobject.Node {
	nodesAfterNodeStatus := f.checkNodeStatus(nodes)
	nodesAfterNodeSelector := f.checkNodeSelector(pod, nodesAfterNodeStatus)
	if len(nodesAfterNodeSelector) == 0 {
		return nodesAfterNodeSelector
	}
	cpuRequest, memoryRequest := f.GetResourceRequest(pod)
	nodesAfterResource := f.checkResource(cpuRequest, memoryRequest, nodesAfterNodeSelector)
	return nodesAfterResource
}
