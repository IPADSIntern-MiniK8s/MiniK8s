package kubeproxy

import (
	"fmt"
	"minik8s/pkg/kubeproxy/ipvs"
)

func Run() {
	ipvs.Init()
	ipvs.TestConfig()
	fmt.Println("end")
	//runLoop()
}

func runLoop() {

}

func syncRunner() {

}
