package pod

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubelet/container"
	"minik8s/pkg/kubelet/utils"
	"os"
	"strconv"
	"strings"
)

// use pointer to add pause container
func CreatePod(pod apiobject.Pod, apiserverAddr string) (bool, string) {
	//ctx := context.Background()
	output, err := utils.Ctl(pod.Data.Namespace, "run", "-d", "--net", "flannel", "--name", generateContainerName("", pod.Data.Name, true), "registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.6")
	//print(output)
	if err != nil {
		fmt.Println(output)
		return false, ""
	}
	pauseContainerID := ""
	outputLen := len(output)
	if outputLen < 100 {
		pauseContainerID = output[:12]
	} else {
		pauseContainerID = output[outputLen-64-1 : outputLen-64+12-1]
	}
	//fmt.Println(pauseContainerID)
	//although in one network namespace, other containers do not have the same network config files as pause, like dns server
	output, err = utils.Ctl(pod.Data.Namespace, "cp", pauseContainerID+":/etc/resolv.conf", "./resolv.conf")
	if err != nil {
		fmt.Println(output)
		return false, ""
	}
	defer os.Remove("./resolv.conf")

	if !addCoreDns("./resolv.conf", apiserverAddr) {
		fmt.Println("add coredns failed")
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
		fmt.Println(err.Error())
		return false, ""
	}
	pid, err := strconv.Atoi(pausePid)
	if err != nil {
		fmt.Println(err.Error())
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
			fmt.Println(err.Error())
			return false, ""
		}
		_, err = utils.Ctl(pod.Data.Namespace, "cp", "./hosts", cSpec.Name+":/etc/hosts")
		if err != nil {
			fmt.Println(err.Error())
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

func DeletePod(pod apiobject.Pod) bool {
	client, err := container.NewClient(pod.Data.Namespace)
	if err != nil {
		fmt.Println(err)
		return false
	}
	ctx := context.Background()
	containers, err := client.Containers(ctx)
	if err != nil {
		return false
	}
	containerNames := map[string]bool{}

	for _, c := range pod.Spec.Containers {
		n := generateContainerName(c.Name, pod.Data.Name, false)
		containerNames[n] = true
		//_, err = utils.Ctl(pod.Data.Namespace, "stop", n)
		//if err != nil {
		//	return false
		//}
		//_, err = utils.Ctl(pod.Data.Namespace, "rm", n)
		//if err != nil {
		//	return false
		//}
	}
	//fmt.Println(containerNames)
	for _, c := range containers {
		//fmt.Println(c.ID())
		//fmt.Println(container.GetContainerStatus(ctx, c))
		if _, ok := containerNames[c.ID()]; ok {
			//fmt.Println(ok, c.ID())
			if success := container.RemoveContainer(ctx, c); !success {
				return false
			}
		}
	}
	//must delete pause at last, otherwise other containers will be stopped

	pauseName := generateContainerName("", pod.Data.Name, true)

	//use containerd api can find container according to name
	//but pause containerd is created by nerdctl instead of containerd, and nerdctl itself maintains a namestore which should be released

	//nerdctl pkg/idutil/containerwalkercontainerwalker.go
	//containers, err = client.Containers(ctx, fmt.Sprintf("labels.%q==%s", "nerdctl/name", pauseName))
	//if err != nil {
	//	return false
	//}
	//if len(containers) < 0 {
	//	return false
	//}
	//if success := container.RemoveContainer(ctx, containers[0]); !success {
	//	return false
	//}

	_, err = utils.Ctl(pod.Data.Namespace, "stop", pauseName)
	if err != nil {
		return false
	}
	_, err = utils.Ctl(pod.Data.Namespace, "rm", pauseName)
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
func belongsToPod(containerName, podName string) bool {
	return strings.HasPrefix(containerName, podName)
}

func addCoreDns(path, apiserverAddr string) bool {
	originalData, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	newData := []byte(fmt.Sprintf("nameserver %s\n", apiserverAddr[:strings.Index(apiserverAddr, ":")]))
	newData = append(newData, originalData...)
	//fmt.Println(string(newData))

	// 将新数据写入文件
	err = os.WriteFile(path, newData, 0644)
	if err != nil {
		return false
	}
	return true
}

func GetPodMetrics(namespace, podName string) *apiobject.PodMetrics {
	client, err := container.NewClient(namespace)
	if err != nil {
		return nil
	}
	ctx := context.Background()
	containers, err := client.Containers(ctx)
	cs := make([]containerd.Container, 0, 0)
	for _, c := range containers {
		if belongsToPod(c.ID(), podName) {
			cs = append(cs, c)
		}
	}
	return container.GetContainersMetrics(cs)
}
