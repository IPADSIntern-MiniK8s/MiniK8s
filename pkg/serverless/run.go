package serverless

import (
	"minik8s/pkg/serverless/autoscaler"
	"minik8s/pkg/serverless/eventfilter"
)

func Run() {
	eventfilter.FunctionSync("functions")
	go autoscaler.PeriodicMetric(10)
}
