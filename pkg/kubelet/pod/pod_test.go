package pod

import (
	"context"
	"fmt"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubelet/utils"
	"testing"
	"time"
)

// nerdctl -n testpod stop $(nerdctl -n testpod ps | grep -v CONTAINER |awk '{print $1}')
// nerdctl -n testpod rm $(nerdctl -n testpod ps -a| grep -v CONTAINER |awk '{print $1}')
func TestPod(t *testing.T) {
	//should start etcd and flannld
	namespace := "testpodns"
	pod := apiobject.Pod{
		Data: apiobject.MetaData{Name: "testpod", Namespace: namespace},
		Spec: apiobject.PodSpec{
			Containers: []apiobject.Container{
				{
					Name:      "c1",
					Image:     "docker.io/mcastelino/nettools",
					Command:   []string{"/root/test_mount/test_network"},
					Env:       []apiobject.Env{{Name: "port", Value: "12345"}},
					Resources: apiobject.Resources{Limits: apiobject.Limit{Cpu: "0.8", Memory: "300Mi"}},
					VolumeMounts: []apiobject.VolumeMounts{
						{
							Name:      "test-volume",
							MountPath: "/root/test_mount",
						},
					},
				},
				{
					Name:      "c2",
					Image:     "docker.io/mcastelino/nettools",
					Command:   []string{"/root/test_mount/test_network"},
					Env:       []apiobject.Env{{Name: "port", Value: "23456"}},
					Resources: apiobject.Resources{Limits: apiobject.Limit{Cpu: "0.8", Memory: "300Mi"}},
					VolumeMounts: []apiobject.VolumeMounts{
						{
							Name:      "test-volume",
							MountPath: "/root/test_mount",
						},
					},
				},
			},
			Volumes: []apiobject.Volumes{
				{
					Name:     "test-volume",
					HostPath: apiobject.HostPath{Path: "/home/test_mount"},
				},
			},
		}}
	success, _ := CreatePod(pod, "192.168.1.13:8080")
	if !success {
		t.Fatalf("create pod failed")
	}

	time.Sleep(time.Second)

	success = utils.CheckCmd(namespace, "testpod-c1", []string{"curl", "127.0.0.1:23456"}, "http connect success")
	if !success {
		t.Fatalf("test localhost network failed")
	}
	success = utils.CheckCmd(namespace, "testpod-c2", []string{"curl", "127.0.0.1:12345"}, "http connect success")
	if !success {
		t.Fatalf("test localhost network failed")
	}
	success = utils.CheckCmd(namespace, "testpod-c1", []string{"ping", "www.baidu.com", "-c", "2"}, "64 bytes from")
	if !success {
		t.Fatalf("test outside network failed")
	}
	success = utils.CheckCmd(namespace, "testpod-c1", []string{"curl", "www.baidu.com"}, "百度一下，你就知道")
	if !success {
		t.Fatalf("test outside network failed")
	}
	success = utils.CheckCmd(namespace, "testpod-c2", []string{"ping", "www.baidu.com", "-c", "2"}, "64 bytes from")
	if !success {
		t.Fatalf("test outside network failed")
	}
	success = utils.CheckCmd(namespace, "testpod-c2", []string{"curl", "www.baidu.com"}, "百度一下，你就知道")
	if !success {
		t.Fatalf("test outside network failed")
	}

	// may get "Shutting down, got signal: Terminated" from pause container, it is a normal behavior

	success = DeletePod(pod)
	if !success {
		t.Fatalf("delete pod failed")
	}
	time.Sleep(time.Second)
	client, err := utils.NewClient(namespace)
	if err != nil {
		t.Fatalf("%v", err)
	}
	ctx := context.Background()
	containers, err := client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(containers) != 0 {
		t.Fatalf("rm containers failed")
	}
}

func TestPodCommunication(t *testing.T) {
	//should start etcd and flannld
	namespace := "testpodns"
	pod1 := apiobject.Pod{
		Data: apiobject.MetaData{Name: "pod1", Namespace: namespace},
		Spec: apiobject.PodSpec{
			Containers: []apiobject.Container{
				{
					Name:    "c",
					Image:   "docker.io/mcastelino/nettools",
					Command: []string{"/root/test_mount/test_network"},
					Env:     []apiobject.Env{{Name: "port", Value: "12345"}},
					VolumeMounts: []apiobject.VolumeMounts{
						{
							Name:      "test-volume",
							MountPath: "/root/test_mount",
						},
					},
				},
			},
			Volumes: []apiobject.Volumes{
				{
					Name:     "test-volume",
					HostPath: apiobject.HostPath{Path: "/home/test_mount"},
				},
			},
		}}
	pod2 := apiobject.Pod{
		Data: apiobject.MetaData{Name: "pod2", Namespace: namespace},
		Spec: apiobject.PodSpec{
			Containers: []apiobject.Container{
				{
					Name:    "c",
					Image:   "docker.io/mcastelino/nettools",
					Command: []string{"/root/test_mount/test_network"},
					Env:     []apiobject.Env{{Name: "port", Value: "23456"}},
					VolumeMounts: []apiobject.VolumeMounts{
						{
							Name:      "test-volume",
							MountPath: "/root/test_mount",
						},
					},
				},
			},
			Volumes: []apiobject.Volumes{
				{
					Name:     "test-volume",
					HostPath: apiobject.HostPath{Path: "/home/test_mount"},
				},
			},
		}}
	success, ip1 := CreatePod(pod1, "192.168.1.13:8080")
	if !success {
		t.Fatalf("create pod1 failed")
	}
	success, ip2 := CreatePod(pod2, "192.168.1.13:8080")
	if !success {
		t.Fatalf("create pod2 failed")
	}
	//ip1, err := kubelet.GetInfo(namespace, "pod1-c", ".NetworkSettings.IPAddress")
	//if err != nil {
	//	t.Fatalf("get pod1 ip failed")
	//}
	//ip2, err := kubelet.GetInfo(namespace, "pod2-c", ".NetworkSettings.IPAddress")
	//if err != nil {
	//	t.Fatalf("get pod2 ip failed")
	//}
	success = utils.CheckCmd(namespace, "pod1-c", []string{"curl", fmt.Sprintf("%s:%s", ip2, pod2.Spec.Containers[0].Env[0].Value)}, "http connect success")
	if !success {
		t.Fatalf("test outside network failed")
	}
	success = utils.CheckCmd(namespace, "pod2-c", []string{"curl", fmt.Sprintf("%s:%s", ip1, pod1.Spec.Containers[0].Env[0].Value)}, "http connect success")
	if !success {
		t.Fatalf("test outside network failed")
	}

	success = DeletePod(pod1)
	if !success {
		t.Fatalf("delete pod failed")
	}
	success = DeletePod(pod2)
	if !success {
		t.Fatalf("delete pod failed")
	}
	time.Sleep(time.Second)
	client, err := utils.NewClient(namespace)
	if err != nil {
		t.Fatalf("%v", err)
	}
	ctx := context.Background()
	containers, err := client.Containers(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(containers) != 0 {
		t.Fatalf("rm containers failed")
	}
}

func TestPodMetrics(t *testing.T) {
	//should start etcd and flannld
	namespace := "testpodns"
	podName := "testpod"
	pod := apiobject.Pod{
		Data: apiobject.MetaData{Name: podName, Namespace: namespace},
		Spec: apiobject.PodSpec{
			Containers: []apiobject.Container{
				{
					Name:      "c1",
					Image:     "ubuntu",
					Command:   []string{"/root/test_mount/test_cpu"},
					Resources: apiobject.Resources{Limits: apiobject.Limit{Cpu: "0.5"}},
					VolumeMounts: []apiobject.VolumeMounts{
						{
							Name:      "test-volume",
							MountPath: "/root/test_mount",
						},
					},
				},
				{
					Name:    "c2",
					Image:   "ubuntu",
					Command: []string{"/root/test_mount/test_memory"},
					VolumeMounts: []apiobject.VolumeMounts{
						{
							Name:      "test-volume",
							MountPath: "/root/test_mount",
						},
					},
				},
			},
			Volumes: []apiobject.Volumes{
				{
					Name:     "test-volume",
					HostPath: apiobject.HostPath{Path: "/home/test_mount"},
				},
			},
		}}
	success, _ := CreatePod(pod, "192.168.1.13:8080")
	if !success {
		t.Fatalf("create pod failed")
	}
	defer DeletePod(pod)
	time.Sleep(time.Second)
	metrics := GetPodMetrics(namespace, podName)
	if metrics == nil {
		t.Fatalf("get pod metrics failed")
	}
	for _, c := range metrics.Containers {
		if c.Name == podName+"-c1" {
			cpu := c.Usage[apiobject.ResourceCPU]
			if cpu > 600 || cpu <= 400 {
				t.Fatalf("cpu metric error,%v", cpu)
			}
		} else {
			memory := c.Usage[apiobject.ResourceMemory]
			if memory < 200000000 {
				t.Fatalf("memory metric error,%v", memory)
			}
		}
	}
}
