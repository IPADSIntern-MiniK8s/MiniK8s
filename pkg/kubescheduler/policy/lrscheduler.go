package policy

import (
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubescheduler/filter"
)

// LeastRequestScheduler support select node that has least scheduled pods
type LeastRequestScheduler struct {
	Filter    filter.TemplateFilter
	frequency map[string]int
}

func NewLeastRequestScheduler(filter filter.TemplateFilter) *LeastRequestScheduler {
	frq := make(map[string]int)
	return &LeastRequestScheduler{
		frequency: frq,
		Filter:    filter,
	}
}

func (s *LeastRequestScheduler) Schedule(pod *apiobject.Pod, nodes []*apiobject.Node) []*apiobject.Node {
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

	// sort the node by their score, from low to high
	for i := 0; i < length; i++ {
		for j := i + 1; j < length; j++ {
			if scores[i] > scores[j] {
				scores[i], scores[j] = scores[j], scores[i]
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
	if length > 0 {
		s.frequency[nodes[0].Data.Name] += 1
	}
	return nodes
}

func (s *LeastRequestScheduler) Score(node *apiobject.Node) float64 {
	score, ok := s.frequency[node.Data.Name]
	if !ok {
		score = 0
	}
	return float64(score)
}
