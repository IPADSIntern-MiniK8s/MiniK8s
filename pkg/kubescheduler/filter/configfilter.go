package filter

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"regexp"
	"strconv"
	"strings"
)

// ConfigFilter is a filter that filter node by config information
type ConfigFilter struct {
	Name string
}

func NewConfigFilter() *ConfigFilter {
	return &ConfigFilter{
		Name: "ConfigFilter",
	}
}

// PreFilter runs a set of functions against a pod. If any of the functions returns an error, the pod is rejected.
func (f ConfigFilter) PreFilter(pod *apiobject.Pod) bool {
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
func (f ConfigFilter) checkNodeStatus(nodes []*apiobject.Node) []*apiobject.Node {
	// check whether the node is ready and can be found
	result := make([]*apiobject.Node, 0)
	for _, node := range nodes {
		if node.Status.Conditions[0].Status == apiobject.Ready && node.Spec.Unschedulable == false {
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

// CheckNodeStatus wrap checkNodeStatus, only for test, if not test, please comment it
func (f ConfigFilter) CheckNodeStatus(nodes []*apiobject.Node) []*apiobject.Node {
	return f.checkNodeStatus(nodes)
}

// checkNodeSelector checks whether the node has the label that the pod needs
func (f ConfigFilter) checkNodeSelector(pod *apiobject.Pod, nodes []*apiobject.Node) []*apiobject.Node {
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

// CheckNodeSelector wrap checkNodeSelector, only for test, if not test, please comment it
func (f ConfigFilter) CheckNodeSelector(pod *apiobject.Pod, nodes []*apiobject.Node) []*apiobject.Node {
	return f.checkNodeSelector(pod, nodes)
}

// ParseQuantity parses the quantity string to float64
func ParseQuantity(str string) (float64, error) {
	var quantity float64
	var err error

	// 使用正则表达式分离出数值和单位
	re := regexp.MustCompile(`^([\d\.]+)([a-zA-Z]*)$`)
	matches := re.FindStringSubmatch(str)

	if len(matches) == 3 {
		// 将数值部分解析为float64
		quantity, err = strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse quantity: %s", err.Error())
		}

		// 将单位部分转换为标准单位，如将m转换为1/1000
		switch strings.ToLower(matches[2]) {
		case "m":
			quantity /= 1000
		case "k":
			quantity *= 1000
		case "ki":
			quantity *= 1024
		case "mi":
			quantity *= 1024 * 1024
		default:
			log.Info("[ParseQuantity] invalid unit: ", matches[2])
		}

		return quantity, nil
	}

	return 0, fmt.Errorf("invalid quantity string: %s", str)
}

// GetResourceRequest gets the resource that the pod needs
func (f ConfigFilter) GetResourceRequest(pod *apiobject.Pod) (float64, float64) {
	// check whether the pod has resource request
	if pod.Spec.Containers == nil || len(pod.Spec.Containers) == 0 ||
		(pod.Spec.Containers[0].Resources.Requests.Cpu == "" && pod.Spec.Containers[0].Resources.Requests.Memory == "") {
		return 0.0, 0.0
	}

	// calculate the resource that the pod needs
	totalCpu, totalMemory := 0.0, 0.0
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests.Cpu != "" {
			// log.Info("get cpu request: ", container.Resources.Requests.Cpu)
			cpuRequest, err := ParseQuantity(container.Resources.Requests.Cpu)
			if err != nil {
				log.Error("parse cpu request error: ", err)
				continue
			}
			// log.Info("the cpu request is: ", cpuRequest)
			totalCpu += cpuRequest
		}
		if container.Resources.Requests.Memory != "" {
			// log.Info("get memory request: ", container.Resources.Requests.Memory)
			memoryRequest, err := ParseQuantity(container.Resources.Requests.Memory)
			if err != nil {
				log.Error("parse memory request error: ", err)
				continue
			}
			// log.Info("the memory request is: ", memoryRequest)
			totalMemory += memoryRequest
		}
	}
	return totalCpu, totalMemory
}

func (f ConfigFilter) checkResource(cpuRequest float64, memoryRequest float64, nodes []*apiobject.Node) []*apiobject.Node {
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
		cpuAvailable, err := ParseQuantity(cpu)
		if err != nil {
			log.Error("parse cpu error: ", err)
			result = append(result, node)
			continue
		}
		if cpuAvailable < cpuRequest {
			continue
		}

		// check whether the node has enough memory
		memory, ok := node.Status.Allocatable["memory"]
		if !ok {
			result = append(result, node)
			continue
		}
		memoryAvailable, err := ParseQuantity(memory)
		if err != nil {
			log.Error("parse memory error: ", err)
			result = append(result, node)
			continue
		}
		if memoryAvailable < memoryRequest {
			continue
		}

		result = append(result, node)
	}

	return result
}

// CheckResource wrap checkResource, only for test, if not test, please comment it
func (f ConfigFilter) CheckResource(cpuRequest float64, memoryRequest float64, nodes []*apiobject.Node) []*apiobject.Node {
	return f.checkResource(cpuRequest, memoryRequest, nodes)
}

func (f ConfigFilter) Filter(pod *apiobject.Pod, nodes []*apiobject.Node) []*apiobject.Node {
	nodesAfterNodeStatus := f.checkNodeStatus(nodes)
	nodesAfterNodeSelector := f.checkNodeSelector(pod, nodesAfterNodeStatus)
	if len(nodesAfterNodeSelector) == 0 {
		return nodesAfterNodeSelector
	}
	cpuRequest, memoryRequest := f.GetResourceRequest(pod)
	nodesAfterResource := f.checkResource(cpuRequest, memoryRequest, nodesAfterNodeSelector)
	return nodesAfterResource
}
