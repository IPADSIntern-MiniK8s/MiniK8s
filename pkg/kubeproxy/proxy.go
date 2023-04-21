package kubeproxy

import (
	"minik8s/pkg/kubeproxy/ipvs"
)

func Run() {
	ipvs.Init()
	ipvs.TestConfig()
	//runLoop()
}

func runLoop() {

}

func syncRunner() {

}
