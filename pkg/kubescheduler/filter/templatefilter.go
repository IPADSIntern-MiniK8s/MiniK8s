package filter

import "minik8s/pkg/apiobject"

// TemplateFilter is a interface for filter
type TemplateFilter interface {
	// Filter give a list a node and a pod, return a list of node
	Filter(pod *apiobject.Pod, nodes []*apiobject.Node) []*apiobject.Node
	// PreFilter runs a set of functions against a pod. If any of the functions returns an error, the pod is rejected.
	PreFilter(pod *apiobject.Pod) bool
}
