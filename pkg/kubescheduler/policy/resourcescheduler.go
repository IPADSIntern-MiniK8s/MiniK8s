package policy

import (
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubescheduler/filter"
	"minik8s/utils/resourceutils"
	"sort"

	log "github.com/sirupsen/logrus"
)

type ResourceScheduler struct {
	Filter   filter.TemplateFilter
	PodQueue map[string]*apiobject.Pod
}

func NewResourceScheduler(filter filter.TemplateFilter) *ResourceScheduler {
	newQueue := make(map[string]*apiobject.Pod)
	return &ResourceScheduler{
		Filter:   filter,
		PodQueue: newQueue,
	}
}

func (s ResourceScheduler) Schedule(pod *apiobject.Pod, nodes []*apiobject.Node) []*apiobject.Node {
	// first precheck the pod
	ret := s.Filter.PreFilter(pod)
	if !ret {
		return nil
	}

	// then filter the nodes
	nodes = s.Filter.Filter(pod, nodes)

	// sort the node by their score
	length := len(nodes)
	scores := make([]float64, length)
	for i, node := range nodes {
		scores[i] = s.Score(node)
		log.Info("[Schedule] node ", node.Data.Name, " score: ", scores[i])
	}
	sort.Slice(nodes, func(i, j int) bool {
		return scores[i] > scores[j]
	})
	return nodes
}

// Score is according the node's capacity
func (s ResourceScheduler) Score(node *apiobject.Node) float64 {
	totalScore := 0.0
	if node.Status.Allocatable == nil {
		return totalScore
	}

	cpu, ok := node.Status.Allocatable["cpu"]
	cpuCap, capok := node.Status.Capability["cpu"]
	if ok && capok {
		cpuAvailable, _ := resourceutils.ParseQuantity(cpu)
		cpuCap, _ := resourceutils.ParseQuantity(cpuCap)
		totalScore += cpuAvailable / cpuCap
	}

	memory, ok := node.Status.Allocatable["memory"]
	memoryCap, capok := node.Status.Capability["memory"]
	if ok && capok {
		memoryAvailable, _ := resourceutils.ParseQuantity(memory)
		memoryCap, _ := resourceutils.ParseQuantity(memoryCap)
		totalScore += memoryAvailable / memoryCap
	}

	return totalScore
}
