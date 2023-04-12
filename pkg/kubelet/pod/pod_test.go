package pod

import (
	"minik8s/pkg/kubelet/container"
	"testing"
)

func TestCreatePod(t *testing.T) {
	pod := Pod{
		Name: "testpod",
		Containers: []container.ContainerSpec{
			{
				Image: "docker.io/library/ubuntu:latest",
				Name:  "test-container",
				Mounts: map[string]string{
					"/home/test_mount": "/root/test_mount",
				},
				CPU: container.CPUSpec{
					Type:  container.CPUCoreID,
					Value: "1",
				},
				CmdLine: "/root/test_mount/test_cpu", //test_cpu test_memory
			},
		},
	}
	success := CreatePod(pod)
	if !success {
		t.Fatalf("create pod failed")
	}
}
