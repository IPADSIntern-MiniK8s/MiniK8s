package container

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
)

type ContainerSpec struct {
	Image   string
	Name    string
	Mounts  map[string]string
	CPU     CPUSpec
	Memory  uint64
	CmdLine string
}

func CreateContainer(ctx context.Context, spec ContainerSpec) containerd.Container {
	//must add tag and host
	client, err := NewClient()
	if err != nil {
		return nil
	}
	//TODO parse imageName, add latest if necessary
	image, err := client.Pull(ctx, spec.Image, containerd.WithPullUnpack)
	if err != nil {
		return nil
	}
	newContainer, err := client.NewContainer(
		ctx,
		spec.Name, //container name
		containerd.WithNewSnapshot(spec.Name, image), //rootfs?
		containerd.WithNewSpec(oci.WithImageConfig(image),
			GenerateHostnameSpec(spec.Name),
			GenerateMountSpec(spec.Mounts),
			GenerateCPUSpec(spec.CPU),
			GenerateCMDSpec(spec.CmdLine),
			GenerateMemorySpec(spec.Memory),
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
