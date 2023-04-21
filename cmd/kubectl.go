package main

import (
	"fmt"
	"minik8s/pkg/kubectl/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
	}
}
