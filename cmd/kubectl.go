package main

import (
	"fmt"
	"minik8s/pkg/kubectl/cmd"
)

//import "k8s-test/pkg/kubectl"

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
	}
}
