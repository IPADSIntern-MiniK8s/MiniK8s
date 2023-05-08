package policy

import "minik8s/pkg/apiobject"

// 需要的函数
// Filter
// - isNodeSuitable
// Bind: Bind binds a pod to a node
// Score: Score extensions implement scoring functions for Scheduler
// priority: gets a priority function for the custom scheduler
// PreFilter: PreFilter runs a set of functions against a pod. If any of the functions returns an error, the pod is rejected.

type Scheduler interface {
	// Bind Bind binds a pod to a node
	Bind(pod *apiobject.Pod, node *apiobject.Node) error // Bind binds a pod to a node
	// Schedule schedules a pod on a node
	Schedule(pod *apiobject.Pod, nodes []*apiobject.Node) []*apiobject.Node
	// Score extensions implement scoring functions for Scheduler
	Score(node *apiobject.Node) float64
}
