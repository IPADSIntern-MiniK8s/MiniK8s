package container

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	"syscall"
)

type ContainerSpec struct {
	Image              string
	Name               string
	ContainerNamespace string
	Mounts             map[string]string
	CPU                CPUSpec
	Memory             uint64
	CmdLine            []string
	Envs               []string
	LinuxNamespaces    map[string]string
}

func CreateContainer(ctx context.Context, spec ContainerSpec) containerd.Container {
	//must add tag and host
	client, err := NewClient(spec.ContainerNamespace)
	if err != nil {
		return nil
	}
	//TODO if first may outoftime?
	image, err := client.Pull(ctx, PadImageName(spec.Image), containerd.WithPullUnpack)
	if err != nil {
		return nil
	}
	opts := []oci.SpecOpts{oci.WithImageConfig(image), GenerateHostnameSpec(spec.Name)}
	if spec.Mounts != nil && len(spec.Mounts) > 0 {
		opts = append(opts, GenerateMountSpec(spec.Mounts))
	}
	if spec.CPU.Type != CPUNone {
		opts = append(opts, GenerateCPUSpec(spec.CPU))
	}
	if spec.CmdLine != nil {
		opts = append(opts, GenerateCMDSpec(spec.CmdLine))
	}
	if spec.Memory != 0 {
		opts = append(opts, GenerateMemorySpec(spec.Memory))
	}
	if spec.Envs != nil && len(spec.Envs) > 0 {
		opts = append(opts, GenerateEnvSpec(spec.Envs))
	}
	if spec.LinuxNamespaces != nil {
		for nsType, path := range spec.LinuxNamespaces {
			opts = append(opts, GenerateNamespaceSpec(nsType, path))
		}
	}
	newContainer, err := client.NewContainer(
		ctx,
		spec.Name, //container name
		containerd.WithNewSnapshot(spec.Name, image), //rootfs?
		containerd.WithNewSpec(opts...),
	)
	if err != nil {
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
	return task.Pid()
	//status, err := task.Wait(ctx)
}

// copy from nerdctl pkg/cmd/container/remove.go
func RemoveContainer(ctx context.Context, containerToRemove containerd.Container) bool {
	task, err := containerToRemove.Task(ctx, nil)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	status, err := task.Status(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	switch status.Status {
	case containerd.Created, containerd.Stopped:
		if _, err := task.Delete(ctx); err != nil {
			fmt.Println(err.Error())
			return false
		}
		return true
	case containerd.Paused:
		if _, err := task.Delete(ctx, containerd.WithProcessKill); err != nil {
			fmt.Println(err.Error())
			return false
		}
		return true
	default:
		//fmt.Println("default")
		if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
			fmt.Println(err.Error())
			return false
		}
		es, err := task.Wait(ctx)
		if err == nil {
			<-es
		}
		if _, err = task.Delete(ctx, containerd.WithProcessKill); err != nil {
			fmt.Println(err.Error())
			return false
		}
	}
	if err = containerToRemove.Delete(ctx); err != nil {
		return false
	}
	//fmt.Println("success")
	return true
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
