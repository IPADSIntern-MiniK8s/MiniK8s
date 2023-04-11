package container

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
)

func CreateContainer(ctx context.Context, imageName string, containerName string, mounts map[string]string) containerd.Container {
	//must add tag and host
	client, err := NewClient()
	if err != nil {
		return nil
	}
	//TODO parse imageName, add latest if necessary
	image, err := client.Pull(ctx, imageName, containerd.WithPullUnpack)
	if err != nil {
		return nil
	}
	newContainer, err := client.NewContainer(
		ctx,
		containerName, //container name
		containerd.WithNewSnapshot(containerName, image), //rootfs?
		containerd.WithNewSpec(oci.WithImageConfig(image),
			GenerateHostnameSpec(containerName),
			GenerateMountSpec(mounts),
		),
	)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return newContainer

}
func StartContainer(ctx context.Context, containerToStart containerd.Container) uint32 {
	task, err := containerToStart.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		fmt.Println(err)
		return 0
	}
	err = task.Start(ctx)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	task.Wait(ctx)
	return task.Pid()
	//status, err := task.Wait(ctx)
}

//TODO stop

//TODO status
