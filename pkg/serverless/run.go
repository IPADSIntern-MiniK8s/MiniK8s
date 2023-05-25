package serverless

import (
	"minik8s/pkg/serverless/autoscaler"
	"minik8s/pkg/serverless/eventfilter"
)

func Run() {
	go autoscaler.PeriodicMetric(30)
	go eventfilter.FunctionSync("functions")
	go eventfilter.WorkFlowSync("workflowexecutors")
}
