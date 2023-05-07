package container

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"minik8s/pkg/kubelet"
	"testing"
	"time"
)

func TestContainer(t *testing.T) {
	ctx := context.Background()
	spec := ContainerSpec{
		Image:              "docker.io/library/ubuntu:latest",
		Name:               "test-container",
		ContainerNamespace: "test",
		Mounts: map[string]string{
			"/home/test_mount": "/root/test_mount",
		},
		CPU: CPUSpec{
			Type:  CPUCoreID,
			Value: "1",
		},
		Memory:  100 * 1024 * 1024,                     //100M
		CmdLine: []string{"/root/test_mount/test_cpu"}, //test_cpu test_memory
		Envs:    []string{"envA=envAvalue", "envB=envBvalue"},
	}
	kubelet.Ctl(spec.ContainerNamespace, "stop", spec.Name)
	kubelet.Ctl(spec.ContainerNamespace, "rm", spec.Name)
	time.Sleep(time.Second * 2)
	client, err := NewClient(spec.ContainerNamespace)
	if err != nil {
		t.Fatalf("%v", err)
	}
	containers, err := client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(containers) > 0 {
		fmt.Println(GetContainerStatus(ctx, containers[0]))
		t.Fatalf("make sure there is no container created before test")
	}
	container := CreateContainer(ctx, spec)
	if container == nil {
		t.Fatalf("create container failed")
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)
	pid := StartContainer(ctx, container)
	if pid == 0 {
		t.Fatalf("start container failed")
	}
	t.Logf("container started, use htop to see cpu utilization")
	time.Sleep(time.Second * 10)

	hostnameCorrect := kubelet.CheckCmd(spec.ContainerNamespace, spec.Name, []string{"hostname"}, spec.Name)
	if !hostnameCorrect {
		t.Fatalf("hostname set failed")
	}
	envExist := kubelet.CheckCmd(spec.ContainerNamespace, spec.Name, []string{"printenv"}, spec.Envs[0])
	if !envExist {
		t.Fatalf("env set failed")
	}
	containers, err = client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(containers) != 1 {
		t.Fatalf("container status wrong")
	}
	c := containers[0]
	if c.ID() != spec.Name {
		t.Fatalf("wrong container")
	}
	if GetContainerStatus(ctx, c) != "running" {
		t.Fatalf("container status wrong")
	}

	_, err = kubelet.Ctl(spec.ContainerNamespace, "stop", spec.Name)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if GetContainerStatus(ctx, c) != "stopped" {
		t.Fatalf("container status wrong")
	}
	_, err = kubelet.Ctl(spec.ContainerNamespace, "rm", spec.Name)
	if err != nil {
		t.Fatalf("%v", err)
	}
	containers, _ = client.Containers(ctx)
	if len(containers) > 0 {
		t.Fatalf("rm container failed")
	}
}

func TestPadImageName(t *testing.T) {
	answer := "docker.io/library/ubuntu:latest"
	if PadImageName("ubuntu") != answer {
		t.Fatalf("pad image name wrong")
	}
	if PadImageName("ubuntu:latest") != answer {
		t.Fatalf("pad image name wrong")
	}
	if PadImageName("docker.io/library/ubuntu") != answer {
		t.Fatalf("pad image name wrong")
	}
	if PadImageName(answer) != answer {
		t.Fatalf("pad image name wrong")
	}
}

func TestGetContainerStatus(t *testing.T) {
	client, err := NewClient("test")
	if err != nil {
		t.Fatalf("%v", err)
	}
	ctx := context.Background()
	containers, err := client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	for _, c := range containers {
		fmt.Println(c.ID())
		fmt.Println(GetContainerStatus(ctx, c))
	}
}
