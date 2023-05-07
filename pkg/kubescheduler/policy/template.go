package policy

// 需要的函数
// Filter
// - isNodeSuitable
// Bind: Bind binds a pod to a node
// Score: Score extensions implement scoring functions for Scheduler
// priority: gets a priority function for the custom scheduler
// PreFilter: PreFilter runs a set of functions against a pod. If any of the functions returns an error, the pod is rejected.
