package container

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"testing"
	"time"
)

func TestContainer(t *testing.T) {
	ctx := context.Background()
	spec := ContainerSpec{
		Image: "docker.io/library/ubuntu:latest",
		Name:  "test-container",
		Mounts: map[string]string{
			"/home/test_mount": "/root/test_mount",
		},
		CPU: CPUSpec{
			Type:  CPUCoreID,
			Value: "1",
		},
		Memory:  100 * 1024 * 1024,           //100M
		CmdLine: "/root/test_mount/test_cpu", //test_cpu test_memory
		Envs:    []string{"a=b", "c=d"},
	}
	_, _ = ctl("stop", spec.Name)
	_, _ = ctl("rm", spec.Name)
	client, _ := NewClient()
	containers, _ := client.Containers(ctx)
	if len(containers) > 0 {
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

	containers, _ = client.Containers(ctx)
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

	ctl("stop", spec.Name)
	if GetContainerStatus(ctx, c) != "stopped" {
		t.Fatalf("container status wrong")
	}
	ctl("rm", spec.Name)
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
	client, _ := NewClient()
	ctx := context.Background()
	containers, _ := client.Containers(ctx)
	for _, c := range containers {
		fmt.Println(c.ID())
		fmt.Println(GetContainerStatus(ctx, c))
	}
}
