package pod

import (
	"context"
	"fmt"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubelet/container"
	"minik8s/pkg/kubelet/utils"
	"os"
	"strconv"
)

// use pointer to add pause container
func CreatePod(pod apiobject.Pod) (bool, string) {
	//ctx := context.Background()
	output, err := utils.Ctl(pod.Data.Namespace, "run", "-d", "--net", "flannel", "--name", generateContainerName("", pod.Data.Name, true), "registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.6")
	if err != nil {
		fmt.Println(output)
		return false, ""
	}
	pauseContainerID := output[:12]
	//although in one network namespace, other containers do not have the same network config files as pause, like dns server
	_, err = utils.Ctl(pod.Data.Namespace, "cp", pauseContainerID+":/etc/resolv.conf", "./resolv.conf")
	if err != nil {
		fmt.Println(output)
		return false, ""
	}
	defer os.Remove("./resolv.conf")

	if !addCoreDns("./resolv.conf") {
		return false, ""
	}

	_, err = utils.Ctl(pod.Data.Namespace, "cp", pauseContainerID+":/etc/hosts", "./hosts")
	if err != nil {
		fmt.Println(output)
		return false, ""
	}
	defer os.Remove("./hosts")
	pausePid, err := utils.GetInfo(pod.Data.Namespace, pauseContainerID, ".State.Pid")
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	pid, err := strconv.Atoi(pausePid)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	namespacePathPrefix := fmt.Sprintf("/proc/%d/ns/", pid)
	ctx := context.Background()
	for _, apiContainerSpec := range pod.Spec.Containers {
		cSpec := apiContainer2Container(pod.Data, pod.Spec.Volumes, apiContainerSpec, namespacePathPrefix)
		c := container.CreateContainer(ctx, cSpec)
		if c == nil {
			return false, ""
		}
		pid := container.StartContainer(ctx, c)
		if pid == 0 {
			return false, ""
		}
		_, err = utils.Ctl(pod.Data.Namespace, "cp", "./resolv.conf", cSpec.Name+":/etc/resolv.conf")
		if err != nil {
			fmt.Println(err)
			return false, ""
		}
		_, err = utils.Ctl(pod.Data.Namespace, "cp", "./hosts", cSpec.Name+":/etc/hosts")
		if err != nil {
			fmt.Println(err)
			return false, ""
		}
	}
	ip, err := utils.GetInfo(pod.Data.Namespace, pauseContainerID, ".NetworkSettings.IPAddress")

	//add pause container to pod,for deleting
	pod.Spec.Containers = append(pod.Spec.Containers, apiobject.Container{
		Name: generateContainerName(pauseContainerID, pod.Data.Name, true),
	})

	return true, ip
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
		Name:               generateContainerName(apicontainer.Name, metaData.Name, false),
		ContainerNamespace: metaData.Namespace,
		Mounts:             mounts,
		CPU: container.CPUSpec{
			Type:  container.CPUNumber,
			Value: apicontainer.Resources.Limits.Cpu,
		},
		Memory:  memory,
		CmdLine: parseCmd(apicontainer.Command, apicontainer.Args),
		Envs:    parseEnv(apicontainer.Env),
		//github.com/opencontainers/runtime-spec/specs-go/config.go/LinuxNamespaceType
		LinuxNamespaces: map[string]string{
			"pid":     namespacePathPrefix + "pid",
			"network": namespacePathPrefix + "net",
			"ipc":     namespacePathPrefix + "ipc",
			"uts":     namespacePathPrefix + "uts",
		},
	}
	return c
}

/*
1. its easier to create pause by nerdctl than by api (hard to set network)
2. nerdctl can only set containerName, not containerID
3. containerd api can only get ID, not containerName
4. no place to store containerid because pause container belongs to a deep implementation of kubelet itself,which should not be seen by apiserver or kubectl


*/

func DeletePod(pod apiobject.Pod) bool {
	var err error
	for _, c := range pod.Spec.Containers {
		n := generateContainerName(c.Name, pod.Data.Name, false)
		_, err = utils.Ctl(pod.Data.Namespace, "stop", n)
		if err != nil {
			return false
		}
		_, err = utils.Ctl(pod.Data.Namespace, "rm", n)
		if err != nil {
			return false
		}
	}

	//delete pause
	name := generateContainerName("", pod.Data.Name, true)
	_, err = utils.Ctl(pod.Data.Namespace, "stop", name)
	if err != nil {
		return false
	}
	_, err = utils.Ctl(pod.Data.Namespace, "rm", name)
	if err != nil {
		return false
	}
	return true
}

func generateContainerName(containerName string, podName string, isPause bool) string {
	if isPause {
		return podName + "-pause"
	}
	return fmt.Sprintf("%s-%s", podName, containerName)
}

func addCoreDns(path string) bool {
	originalData, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	newData := []byte("nameserver 192.168.1.13\n")
	newData = append(newData, originalData...)
	//fmt.Println(string(newData))

	// 将新数据写入文件
	err = os.WriteFile(path, newData, 0644)
	if err != nil {
		return false
	}
	return true
}
