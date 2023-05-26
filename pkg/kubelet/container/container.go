package container

import (
	"context"
	"fmt"
	v1 "github.com/containerd/cgroups/stats/v1"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	"github.com/gogo/protobuf/proto"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/apiobject/utils"
	"minik8s/pkg/kubelet/image"
	kubeletutils "minik8s/pkg/kubelet/utils"
	"reflect"
	"syscall"
	"time"
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
	Labels             map[string]string
}

func CreateContainer(ctx context.Context, spec ContainerSpec) containerd.Container {
	//must add tag and host
	client, err := kubeletutils.NewClient(spec.ContainerNamespace)
	if err != nil {
		fmt.Println("new client failed")
		return nil
	}

	im := image.EnsureImage(spec.ContainerNamespace, client, spec.Image)
	if im == nil {
		fmt.Println("get image failed")
		return nil
	}
	opts := []oci.SpecOpts{oci.WithImageConfig(im), GenerateHostnameSpec(spec.Name)}
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

	copts := []containerd.NewContainerOpts{
		containerd.WithNewSnapshot(spec.Name, im), //rootfs?
		containerd.WithNewSpec(opts...),
	}
	if spec.Labels!=nil && len(spec.Labels) > 0{
		copts = append(copts,containerd.WithContainerLabels(spec.Labels))
	}
	newContainer, err := client.NewContainer(
		ctx,
		spec.Name, //container name
		copts...
	)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return newContainer

}
func StartContainer(ctx context.Context, containerToStart containerd.Container) uint32 {
	task, err := containerToStart.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		fmt.Printf("Newtask failed:%v\n",err)
		return 0
	}
	err = task.Start(ctx)
	if err != nil {
		fmt.Printf("task start failed:%v\n",err)
		return 0
	}
	return task.Pid()
	//status, err := task.Wait(ctx)
}

// copy from nerdctl pkg/cmd/container/remove.go
func RemoveContainer(ctx context.Context, containerToRemove containerd.Container) bool {
	task, err := containerToRemove.Task(ctx, nil)
	if err == nil {
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
	}
	
	var delOpts []containerd.DeleteOpts
	if _, err := containerToRemove.Image(ctx); err == nil {
		delOpts = append(delOpts, containerd.WithSnapshotCleanup)
	}

	if containerToRemove.Delete(ctx, delOpts...) != nil {
		if containerToRemove.Delete(ctx) != nil {
			return false
		}
	}
	//fmt.Println("success")
	return true
}

func GetContainerStatus(ctx context.Context, c containerd.Container) (string,uint32) {
	task, err := c.Task(ctx, nil)
	if err != nil {
		return err.Error(),0
	}
	status, err := task.Status(ctx)
	if err != nil {
		return err.Error(),0
	}
	return string(status.Status),status.ExitStatus
}

type metricsCollection struct {
	begin       time.Time
	tasks       []containerd.Task
	preTimes    []time.Time
	preCPUs     []uint64
	CPUPercents []uint64
	memorys     []uint64
}

func GetContainersMetrics(cs []containerd.Container) *apiobject.PodMetrics {
	if len(cs) == 0 {
		return &apiobject.PodMetrics{}
	}
	ctx := context.Background()
	collection := metricsCollection{
		begin:       time.Now(),
		tasks:       []containerd.Task{},
		preTimes:    []time.Time{},
		preCPUs:     []uint64{},
		CPUPercents: []uint64{},
		memorys:     []uint64{},
	}
	for _, c := range cs {
		task, err := c.Task(ctx, nil)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		collection.tasks = append(collection.tasks, task)
		collection.preTimes = append(collection.preTimes, time.Now())
		collection.preCPUs = append(collection.preCPUs, 0)
		collection.CPUPercents = append(collection.CPUPercents, 0)
		collection.memorys = append(collection.memorys, 0)
	}

	podMetrics := &apiobject.PodMetrics{Window: utils.Duration{time.Second * 5},
		Containers: []apiobject.ContainerMetrics{},
	}
	//fmt.Println(task.Pid())

	var data v1.Metrics
	var v interface{}
	var curTime time.Time
	var curCPU uint64
	collection.begin = time.Now()
	for i := 0; i <= 1; i++ {
		for ti, task := range collection.tasks {
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
				return nil
			}
			if i == 0 {
				collection.preTimes[ti] = curTime
				collection.preCPUs[ti] = data.CPU.Usage.Total
				collection.memorys[ti] = data.Memory.Usage.Usage
				time.Sleep(podMetrics.Window.Duration)
				continue
			}
			collection.memorys[ti] += data.Memory.Usage.Usage
			collection.memorys[ti] /= 2

			//fmt.Println("memory:", data.Memory.Usage.Usage)
			//fmt.Println("CPU:", data.CPU.Usage.Total)
			curCPU = data.CPU.Usage.Total
			cpuDelta := curCPU - collection.preCPUs[ti]

			timeDelta := curTime.Sub(collection.preTimes[ti])
			collection.CPUPercents[ti] = uint64(float64(cpuDelta) / float64(timeDelta.Nanoseconds()) * 1000)
		}
	}
	podMetrics.Timestamp.Time = curTime
	for ci, c := range cs {
		podMetrics.Containers = append(podMetrics.Containers, apiobject.ContainerMetrics{
			Name: c.ID(),
			Usage: map[apiobject.ResourceName]utils.Quantity{
				apiobject.ResourceCPU:    utils.Quantity(collection.CPUPercents[ci]),
				apiobject.ResourceMemory: utils.Quantity(collection.memorys[ci]),
			},
		})
	}
	return podMetrics
}
