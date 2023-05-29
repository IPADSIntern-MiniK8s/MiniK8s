package container

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubelet/utils"
	"testing"
	"time"
)

func TestContainer(t *testing.T) {
	ctx := context.Background()
	spec := ContainerSpec{
		Image:              "docker.io/library/ubuntu:latest",
		Name:               "test-container",
		ContainerNamespace: "test",
		Mounts: map[string]string{
			"/home/test_mount": "/root/test_mount",
		},
		CPU: CPUSpec{
			Type:  CPUCoreID,
			Value: "1",
		},
		Memory:  100 * 1024 * 1024,                     //100M
		CmdLine: []string{"/root/test_mount/test_cpu"}, //test_cpu test_memory
		Envs:    []string{"envA=envAvalue", "envB=envBvalue"},
	}
	utils.Ctl(spec.ContainerNamespace, "stop", spec.Name)
	utils.Ctl(spec.ContainerNamespace, "rm", spec.Name)
	time.Sleep(time.Second * 2)
	client, err := utils.NewClient(spec.ContainerNamespace)
	if err != nil {
		t.Fatalf("%v", err)
	}
	containers, err := client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(containers) > 0 {
		fmt.Println(GetContainerStatus(ctx, containers[0]))
		t.Fatalf("make sure there is no container created before test")
	}
	container := CreateContainer(ctx, spec)
	if container == nil {
		t.Fatalf("create container failed")
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)
	pid := StartContainer(ctx, container)
	if pid == 0 {
		t.Fatalf("start container failed")
	}
	t.Logf("container started, use htop to see cpu utilization")
	time.Sleep(time.Second * 10)

	hostnameCorrect := utils.CheckCmd(spec.ContainerNamespace, spec.Name, []string{"hostname"}, spec.Name)
	if !hostnameCorrect {
		t.Fatalf("hostname set failed")
	}
	envExist := utils.CheckCmd(spec.ContainerNamespace, spec.Name, []string{"printenv"}, spec.Envs[0])
	if !envExist {
		t.Fatalf("env set failed")
	}
	containers, err = client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(containers) != 1 {
		t.Fatalf("container status wrong")
	}
	c := containers[0]
	if c.ID() != spec.Name {
		t.Fatalf("wrong container")
	}
	if s,_:=GetContainerStatus(ctx, c);s != "running" {
		t.Fatalf("container status wrong")
	}

	_, err = utils.Ctl(spec.ContainerNamespace, "stop", spec.Name)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if s,_:=GetContainerStatus(ctx, c);s != "stopped" {
		t.Fatalf("container status wrong")
	}
	_, err = utils.Ctl(spec.ContainerNamespace, "rm", spec.Name)
	if err != nil {
		t.Fatalf("%v", err)
	}
	containers, _ = client.Containers(ctx)
	if len(containers) > 0 {
		t.Fatalf("rm container failed")
	}
}

func TestRemoveContainer(t *testing.T) {
	ctx := context.Background()
	spec := ContainerSpec{
		Image:              "docker.io/library/ubuntu:latest",
		Name:               "test-container1",
		ContainerNamespace: "test",
		Mounts: map[string]string{
			"/home/test_mount": "/root/test_mount",
		},
		CmdLine: []string{"/root/test_mount/test_network"},
		Envs:    []string{"port=12345"},
	}
	utils.Ctl(spec.ContainerNamespace, "stop", spec.Name)
	utils.Ctl(spec.ContainerNamespace, "rm", spec.Name)
	time.Sleep(time.Second * 2)
	client, err := utils.NewClient(spec.ContainerNamespace)
	if err != nil {
		t.Fatalf("%v", err)
	}
	containers, err := client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(containers) > 0 {
		fmt.Println(GetContainerStatus(ctx, containers[0]))
		t.Fatalf("make sure there is no container created before test")
	}
	container := CreateContainer(ctx, spec)
	if container == nil {
		t.Fatalf("create container failed")
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)
	pid := StartContainer(ctx, container)
	if pid == 0 {
		t.Fatalf("start container failed")
	}
	time.Sleep(time.Second)

	containers, err = client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(containers) != 1 {
		t.Fatalf("container status wrong")
	}
	c := containers[0]
	if c.ID() != spec.Name {
		t.Fatalf("wrong container")
	}
	if s,_:=GetContainerStatus(ctx, c);s != "running" {
		t.Fatalf("container status wrong")
	}

	removed := RemoveContainer(ctx, c)
	if !removed {
		t.Fatalf("remove container failed")
	}
	time.Sleep(time.Second)
	client, err = utils.NewClient(spec.ContainerNamespace)
	containers, _ = client.Containers(ctx)
	if len(containers) > 0 {
		s,_:= GetContainerStatus(ctx, c)
		t.Fatalf("rm container failed,%v:%v", c.ID(),s)
	}
}

func TestGetContainerStatus(t *testing.T) {
	client, err := utils.NewClient("teststatus")
	if err != nil {
		t.Fatalf("%v", err)
	}
	ctx := context.Background()
	containers, err := client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	for _, c := range containers {
		fmt.Println(c.ID())
		fmt.Println(GetContainerStatus(ctx, c))
	}
}

func TestRemoveOneContainer(t *testing.T) {
	client, err := utils.NewClient("testpod1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	ctx := context.Background()
	containers, err := client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(containers) > 0 {
		RemoveContainer(ctx, containers[0])
	}

}

func TestContainerFilter(t *testing.T) {
	ctx := context.Background()
	ns:="filter"
	spec := ContainerSpec{
		Image:              "ubuntu",
		Name:               "test-filter",
		ContainerNamespace: ns,
		CmdLine: []string{"sleep","10"},
		Labels: map[string]string{"pod":"podname"},
	}
	container := CreateContainer(ctx, spec)
	if container == nil {
		t.Fatalf("create container failed")
	}

	client, err := utils.NewClient(ns)
	if err != nil {
		t.Fatalf("%v", err)
	}
	containers, err := client.Containers(ctx, fmt.Sprintf("labels.%q==%s", "pod", "podname"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(containers) !=1{
		t.Fatalf("get container error")
	}
	containers, err = client.Containers(ctx, fmt.Sprintf("labels.%q==%s", "pod", "wrong"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(containers) !=0{
		t.Fatalf("get container error")
	}
	utils.Ctl(spec.ContainerNamespace, "stop", spec.Name)
	utils.Ctl(spec.ContainerNamespace, "rm", spec.Name)

}

func TestContainerMetric(t *testing.T) {
	client, err := utils.NewClient("test-metrics")
	if err != nil {
		t.Fatalf("%v", err)
	}
	ctx := context.Background()
	containers, err := client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(containers) == 0 {
		return
	}
	metrics := GetContainersMetrics(containers)
	if metrics == nil {
		t.Fatalf("get metrics failed")
	}
	for _, mi := range metrics.Containers {
		fmt.Println(mi.Name)
		fmt.Println(mi.Usage[apiobject.ResourceCPU])
		fmt.Println(mi.Usage[apiobject.ResourceMemory])
	}
}
func TestUseLocalImage(t *testing.T) {
	ctx := context.Background()
	spec := ContainerSpec{
		Image:              "master:5000/gpu-server:latest",
		Name:               "test-local-image",
		ContainerNamespace: "default",
		Mounts: map[string]string{
			"/home/test_mount": "/root/test_mount",
		},
		CmdLine: []string{"sleep", "100"}, //test_cpu test_memory
	}
	utils.Ctl(spec.ContainerNamespace, "stop", spec.Name)
	utils.Ctl(spec.ContainerNamespace, "rm", spec.Name)
	//time.Sleep(time.Second * 2)
	container := CreateContainer(ctx, spec)
	if container == nil {
		t.Fatalf("create container failed")
	}
	utils.Ctl(spec.ContainerNamespace, "stop", spec.Name)
	utils.Ctl(spec.ContainerNamespace, "rm", spec.Name)
}
