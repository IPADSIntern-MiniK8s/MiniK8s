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
	Envs    []string
}

func CreateContainer(ctx context.Context, spec ContainerSpec) containerd.Container {
	//must add tag and host
	client, err := NewClient()
	if err != nil {
		return nil
	}
	image, err := client.Pull(ctx, spec.Image, containerd.WithPullUnpack)
	if err != nil {
		return nil
	}
	opts := []oci.SpecOpts{oci.WithImageConfig(image), GenerateHostnameSpec(spec.Name)}
	if len(spec.Mounts) > 0 {
		opts = append(opts, GenerateMountSpec(spec.Mounts))
	}
	if spec.CPU.Type != CPUNone {
		opts = append(opts, GenerateCPUSpec(spec.CPU))
	}
	if spec.CmdLine != "" {
		opts = append(opts, GenerateCMDSpec(spec.CmdLine))
	}
	if spec.Memory != 0 {
		opts = append(opts, GenerateMemorySpec(spec.Memory))
	}
	if len(spec.Envs) > 0 {
		opts = append(opts, GenerateEnvSpec(spec.Envs))
	}
	newContainer, err := client.NewContainer(
		ctx,
		spec.Name, //container name
		containerd.WithNewSnapshot(spec.Name, image), //rootfs?
		containerd.WithNewSpec(opts...),
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

func GetContainerStatus(ctx context.Context, c containerd.Container) string {
	task, err := c.Task(ctx, nil)
	if err != nil {
		return err.Error()
	}
	status, err := task.Status(ctx)
	if err != nil {
		return err.Error()
	}
	return string(status.Status)
}
