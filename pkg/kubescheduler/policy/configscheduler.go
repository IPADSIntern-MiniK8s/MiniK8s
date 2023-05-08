package policy

import (
	"errors"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubescheduler/filter"
	"strconv"
)

type ConfigScheduler struct {
	Filter filter.TemplateFilter
}

func NewConfigScheduler(filter filter.TemplateFilter) *ConfigScheduler {
	return &ConfigScheduler{
		Filter: filter,
	}
}

func (s *ConfigScheduler) Schedule(pod *apiobject.Pod, nodes []*apiobject.Node) []*apiobject.Node {
	// first precheck the pod
	ret := s.Filter.PreFilter(pod)
	if !ret {
		return nil
	}

	// then filter the nodes
	nodes = s.Filter.Filter(pod, nodes)
	return nodes
}

func (s *ConfigScheduler) Bind(pod *apiobject.Pod, node *apiobject.Node) error {
	hostIp := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == "InternalIP" {
			hostIp = addr.Address
		}
	}
	if hostIp == "" {
		return errors.New("node's InternalIP can't be found")
	}
	pod.Status.HostIp = hostIp
	return nil
}

// Score is according the node's capacity
func (s *ConfigScheduler) Score(node *apiobject.Node) float64 {
	totalScore := 0.0
	if node.Status.Allocatable == nil {
		return totalScore
	}

	cpu, ok := node.Status.Allocatable["cpu"]
	cpuCap, capok := node.Status.Capability["cpu"]
	if ok && capok {
		cpuAvailable, _ := strconv.ParseFloat(cpu, 64)
		cpuCap, _ := strconv.ParseFloat(cpuCap, 64)
		totalScore += 1 - cpuAvailable/cpuCap
	}

	memory, ok := node.Status.Allocatable["memory"]
	memoryCap, capok := node.Status.Capability["memory"]
	if ok && capok {
		memoryAvailable, _ := strconv.ParseFloat(memory, 64)
		memoryCap, _ := strconv.ParseFloat(memoryCap, 64)
		totalScore += 1 - memoryAvailable/memoryCap
	}

	return totalScore
}
