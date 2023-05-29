package kubeapiserver

import (
	"minik8s/pkg/kubeapiserver/apimachinery"
)

func Run() {
	myAPI := apimachinery.NewAPI()
	go apimachinery.HeartBeat()
	err := myAPI.Run(":8080")
	if err != nil {
		panic(err)
	}
}
