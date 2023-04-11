package container

import (
	"context"
	"github.com/containerd/containerd"
	"os"
	"testing"
)

func TestCreateContainer(t *testing.T) {
	ctx := context.Background()
	name := "python-container"
	_, _ = ctl("stop", name)
	_, _ = ctl("rm", name)

	mounts := map[string]string{
		"/home/test_mount": "/root/test_mount",
	}
	file, err := os.OpenFile("/home/test_mount/test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf(err.Error())
	}
	mountFileContent := "test mount\n"
	_, err = file.WriteString(mountFileContent)
	if err != nil {
		t.Fatalf(err.Error())
	}

	container := CreateContainer(ctx, "docker.io/library/python:latest", name, mounts)
	if container == nil {
		t.Fatalf("create container failed")
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)
	pid := StartContainer(ctx, container)
	if pid == 0 {
		t.Fatalf("start container failed")
	}

	//no console,test by hand
	//output, err := ctl("exec", "-it", name, "cat", "/root/test_mount/test.txt")

	//fmt.Println(output)

}
