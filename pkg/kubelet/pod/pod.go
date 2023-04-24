package pod

import (
	"context"
	"fmt"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubelet"
	"minik8s/pkg/kubelet/container"
	"os"
	"strconv"
)

func CreatePod(pod apiobject.Pod) bool {
	//ctx := context.Background()
	output, err := kubelet.Ctl(pod.Data.Namespace, "run", "-d", "--net", "flannel", "registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.6")
	if err != nil {
		return false
	}
	pauseContainerID := output[:12]
	//although in one network namespace, other containers do not have the same network config files as pause, like dns server
	_, err = kubelet.Ctl(pod.Data.Namespace, "cp", pauseContainerID+":/etc/resolv.conf", "./resolv.conf")
	if err != nil {
		return false
	}
	defer os.Remove("./resolv.conf")
	_, err = kubelet.Ctl(pod.Data.Namespace, "cp", pauseContainerID+":/etc/hosts", "./hosts")
	if err != nil {
		return false
	}
	defer os.Remove("./hosts")
	pausePid, err := kubelet.Ctl(pod.Data.Namespace, "inspect", "-f", "{{.State.Pid}}", pauseContainerID)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(pausePid[:len(pausePid)-1]) //last is \n
	if err != nil {
		return false
	}
	namespacePathPrefix := fmt.Sprintf("/proc/%d/ns/", pid)
	ctx := context.Background()
	for _, apiContainerSpec := range pod.Spec.Containers {
		cSpec := apiContainer2Container(pod.Data, pod.Spec.Volumes, apiContainerSpec, namespacePathPrefix)
		c := container.CreateContainer(ctx, cSpec)
		if c == nil {
			return false
		}
		pid := container.StartContainer(ctx, c)
		if pid == 0 {
			return false
		}
		_, err = kubelet.Ctl(pod.Data.Namespace, "cp", "./resolv.conf", cSpec.Name+":/etc/resolv.conf")
		if err != nil {
			return false
		}
		_, err = kubelet.Ctl(pod.Data.Namespace, "cp", "./hosts", cSpec.Name+":/etc/hosts")
		if err != nil {
			return false
		}
	}
	return true
}

func apiContainer2Container(metaData apiobject.MetaData, volumes []apiobject.Volumes, apicontainer apiobject.Container, namespacePathPrefix string) container.ContainerSpec {
	mounts := make(map[string]string)
	if apicontainer.VolumeMounts != nil {
		for _, containerV := range apicontainer.VolumeMounts {
			for _, hostV := range volumes {
				if containerV.Name == hostV.Name {
					mounts[hostV.HostPath.Path] = containerV.MountPath
				}
			}
		}
	}
	//kubectl and apiserver should make sure the request is valid
	memory, _ := parseMemory(apicontainer.Resources.Limits.Memory)
	cmd := make([]string, len(apicontainer.Command)+len(apicontainer.Args))
	copy(cmd, apicontainer.Command)
	cmd = append(cmd, apicontainer.Args...)
	c := container.ContainerSpec{
		Image:              apicontainer.Image,
		Name:               fmt.Sprintf("%s-%s", metaData.Name, apicontainer.Name),
		ContainerNamespace: metaData.Namespace,
		Mounts:             mounts,
		CPU: container.CPUSpec{
			Type:  container.CPUNumber,
			Value: apicontainer.Resources.Limits.Cpu,
		},
		Memory:  memory,
		CmdLine: parseCmd(apicontainer.Command, apicontainer.Args),
		Envs:    parseEnv(apicontainer.Env),
		LinuxNamespaces: map[string]string{
			"network": namespacePathPrefix + "net",
		},
	}
	return c
}