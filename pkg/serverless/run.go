package serverless

import (
	"minik8s/pkg/serverless/autoscaler"
	"minik8s/pkg/serverless/eventfilter"
)

func Run() {
	go eventfilter.Sync("functions")
	go autoscaler.PeriodicMetric(10)
}