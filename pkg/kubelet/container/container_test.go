package container

import (
	"context"
	"fmt"
	v1 "github.com/containerd/cgroups/stats/v1"
	"github.com/containerd/containerd"
	"github.com/gogo/protobuf/proto"
	"minik8s/pkg/kubelet/utils"
	"reflect"
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
	client, err := NewClient(spec.ContainerNamespace)
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
	if GetContainerStatus(ctx, c) != "running" {
		t.Fatalf("container status wrong")
	}

	_, err = utils.Ctl(spec.ContainerNamespace, "stop", spec.Name)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if GetContainerStatus(ctx, c) != "stopped" {
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
	client, err := NewClient(spec.ContainerNamespace)
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
	if GetContainerStatus(ctx, c) != "running" {
		t.Fatalf("container status wrong")
	}

	removed := RemoveContainer(ctx, c)
	if !removed {
		t.Fatalf("remove container failed")
	}
	time.Sleep(time.Second)
	client, err = NewClient(spec.ContainerNamespace)
	containers, _ = client.Containers(ctx)
	if len(containers) > 0 {
		t.Fatalf("rm container failed,%v:%v", c.ID(), GetContainerStatus(ctx, c))
	}
}

func TestPadImageName(t *testing.T) {
	answer := "docker.io/library/ubuntu:latest"
	if PadImageName("ubuntu") != answer {
		t.Fatalf("pad image name wrong")
	}
	if PadImageName("ubuntu:latest") != answer {
		t.Fatalf("pad image name wrong")
	}
	if PadImageName("docker.io/library/ubuntu") != answer {
		t.Fatalf("pad image name wrong")
	}
	if PadImageName(answer) != answer {
		t.Fatalf("pad image name wrong")
	}
}

func TestGetContainerStatus(t *testing.T) {
	client, err := NewClient("testpod18")
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
	client, err := NewClient("testpod1")
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
	client, err := NewClient("default")
	if err != nil {
		t.Fatalf("%v", err)
	}
	ctx := context.Background()
	containers, err := client.Containers(ctx, fmt.Sprintf("labels.%q==%s", "nerdctl/name", "testpod2-pause"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	for _, c := range containers {
		fmt.Println(c.ID())
	}
}

func TestContainerMetric(t *testing.T) {
	client, err := NewClient("default")
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
	c := containers[0]
	fmt.Println(c.ID())
	task, err := c.Task(ctx, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(task.Pid())
	var data v1.Metrics
	var v interface{}

	var preTime time.Time
	var curTime time.Time
	var preCPU uint64 = 0
	var curCPU uint64 = 0
	for i := 0; i < 5; i++ {
		metrics, err := task.Metrics(ctx)
		if err != nil {
			continue
		}
		curTime = time.Now()

		//can not use Unmarshall, strange
		//fmt.Println(typeurl.UnmarshalAny(metrics.Data))
		//fmt.Println(typeurl.UnmarshalByTypeURL(metrics.Data.TypeUrl, metrics.Data.Value))
		v = reflect.New(reflect.TypeOf(data)).Interface()
		err = proto.Unmarshal(metrics.Data.Value, v.(proto.Message))
		if err != nil {
			fmt.Println(err.Error())
		}
		switch value := v.(type) {
		case *v1.Metrics:
			data = *value
		default:
			return
		}
		if preCPU == 0 {
			preTime = curTime
			preCPU = data.CPU.Usage.Total
			continue
		}

		fmt.Println("memory:", data.Memory.Usage.Usage)
		fmt.Println("CPU:", data.CPU.Usage.Total)
		curCPU = data.CPU.Usage.Total
		cpuDelta := curCPU - preCPU

		timeDelta := curTime.Sub(preTime)
		cpuPercent := float64(cpuDelta) / float64(timeDelta.Nanoseconds()) * 100.0
		fmt.Println("cpuPercent:", cpuPercent)
		preCPU = data.CPU.Usage.Total
		preTime = curTime
		time.Sleep(time.Second)

	}

}
