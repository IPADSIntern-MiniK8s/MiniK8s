package container

import (
	"context"
	"github.com/containerd/containerd"
	"testing"
)

func TestCreateContainer(t *testing.T) {
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
	}
	_, _ = ctl("stop", spec.Name)
	_, _ = ctl("rm", spec.Name)

	container := CreateContainer(ctx, spec)
	if container == nil {
		t.Fatalf("create container failed")
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)
	pid := StartContainer(ctx, container)
	if pid == 0 {
		t.Fatalf("start container failed")
	}

}
