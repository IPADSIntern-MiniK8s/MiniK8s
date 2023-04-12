package pod

import (
	"context"
	"fmt"
	"minik8s/pkg/kubelet/container"
)

type Pod struct {
	Name       string
	Containers []container.ContainerSpec
}

func CreatePod(pod Pod) bool {
	ctx := context.Background()
	pauseSpec := container.ContainerSpec{
		Image: "registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.6",
		Name:  fmt.Sprintf("%s-%s", pod.Name, "pause"),
	}
	pauseContainer := container.CreateContainer(ctx, pauseSpec)
	if pauseContainer == nil {
		return false
	}
	pid := container.StartContainer(ctx, pauseContainer)
	if pid == 0 {
		return false
	}
	networkPath := fmt.Sprintf("/proc/%d/ns/net", pid)
	for _, cSpec := range pod.Containers {
		if cSpec.Namespaces != nil {
			cSpec.Namespaces["network"] = networkPath
		} else {
			cSpec.Namespaces = map[string]string{
				"network": networkPath,
			}
		}
		cSpec.Name = fmt.Sprintf("%s-%s", pod.Name, cSpec.Name)
		c := container.CreateContainer(ctx, cSpec)
		if c == nil {
			return false
		}
		pid = container.StartContainer(ctx, c)
		if pid == 0 {
			return false
		}
	}
	return true
}
