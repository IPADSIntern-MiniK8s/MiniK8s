package kubeapiserver

import (
	"minik8s/pkg/kubeapiserver/apimachinery"
)

func main() {
	myAPI := apimachinery.NewAPI()
	err := myAPI.Run(":8080")
	if err != nil {
		panic(err)
	}
}
