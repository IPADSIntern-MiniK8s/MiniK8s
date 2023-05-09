package main

import (
	"minik8s/pkg/kubescheduler"
	"os"
)

func main() {
	// use parameter to get the policy
	policy := "default"
	if len(os.Args) > 1 {
		policy = os.Args[1]
	}
	kubescheduler.Run(policy)
}
