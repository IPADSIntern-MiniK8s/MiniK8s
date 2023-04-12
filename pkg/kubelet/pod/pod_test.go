package pod

import (
	"fmt"
	"minik8s/pkg/kubelet"
	"minik8s/pkg/kubelet/container"
	"testing"
)

func TestCreatePod(t *testing.T) {
	pod := Pod{
		Name: "testpod",
		Containers: []container.ContainerSpec{
			{
				Image: "docker.io/mcastelino/nettools:latest",
				Name:  "c1",
				Mounts: map[string]string{
					"/home/test_mount": "/root/test_mount",
				},
				CmdLine: "/root/test_mount/test_network",
				Envs:    []string{"port=12345"},
			},
			{
				Image: "docker.io/mcastelino/nettools:latest",
				Name:  "c2",
				Mounts: map[string]string{
					"/home/test_mount": "/root/test_mount",
				},
				CmdLine: "/root/test_mount/test_network",
				Envs:    []string{"port=23456"},
			},
		},
	}
	success := CreatePod(pod)
	if !success {
		t.Fatalf("create pod failed")
	}

	success = kubelet.CheckCmd("testpod-c1", []string{"curl", "127.0.0.1:23456"}, "http connect success")
	if !success {
		t.Fatalf("test network failed")
	}
	success = kubelet.CheckCmd("testpod-c2", []string{"curl", "127.0.0.1:12345"}, "http connect success")
	if !success {
		t.Fatalf("test network failed")
	}
	//
	for _, c := range pod.Containers {
		n := fmt.Sprintf("%s-%s", pod.Name, c.Name)
		kubelet.Ctl("stop", n)
		kubelet.Ctl("rm", n)
	}
	// may get "Shutting down, got signal: Terminated" from pause container, it is a normal behavior
	kubelet.Ctl("stop", pod.Name+"-pause")
	kubelet.Ctl("rm", pod.Name+"-pause")

}
