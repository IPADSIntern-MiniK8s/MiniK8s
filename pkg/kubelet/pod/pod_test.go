package pod

import (
	"context"
	"fmt"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubelet"
	"minik8s/pkg/kubelet/container"
	"testing"
)

// nerdctl -n testpod stop $(nerdctl -n testpod ps | grep -v CONTAINER |awk '{print $1}')
// nerdctl -n testpod rm $(nerdctl -n testpod ps -a| grep -v CONTAINER |awk '{print $1}')
func TestCreatePod(t *testing.T) {
	//should start etcd and flannld
	namespace := "testpod"
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
	success := CreatePod(pod)
	if !success {
		t.Fatalf("create pod failed")
	}

	success = kubelet.CheckCmd(namespace, "testpod-c1", []string{"curl", "127.0.0.1:23456"}, "http connect success")
	if !success {
		t.Fatalf("test localhost network failed")
	}
	success = kubelet.CheckCmd(namespace, "testpod-c2", []string{"curl", "127.0.0.1:12345"}, "http connect success")
	if !success {
		t.Fatalf("test localhost network failed")
	}
	success = kubelet.CheckCmd(namespace, "testpod-c1", []string{"ping", "www.baidu.com", "-c", "2"}, "64 bytes from")
	if !success {
		t.Fatalf("test outside network failed")
	}
	success = kubelet.CheckCmd(namespace, "testpod-c1", []string{"curl", "www.baidu.com"}, "百度一下，你就知道")
	if !success {
		t.Fatalf("test outside network failed")
	}
	success = kubelet.CheckCmd(namespace, "testpod-c2", []string{"ping", "www.baidu.com", "-c", "2"}, "64 bytes from")
	if !success {
		t.Fatalf("test outside network failed")
	}
	success = kubelet.CheckCmd(namespace, "testpod-c2", []string{"curl", "www.baidu.com"}, "百度一下，你就知道")
	if !success {
		t.Fatalf("test outside network failed")
	}
	//
	for _, c := range pod.Spec.Containers {
		n := fmt.Sprintf("%s-%s", pod.Data.Name, c.Name)
		kubelet.Ctl(namespace, "stop", n)
		kubelet.Ctl(namespace, "rm", n)
	}
	// may get "Shutting down, got signal: Terminated" from pause container, it is a normal behavior
	client, _ := container.NewClient(namespace)
	ctx := context.Background()
	containers, _ := client.Containers(ctx)
	if len(containers) != 1 { //left pause
		t.Fatalf("rm containers failed")
	}
	id := containers[0].ID()
	kubelet.Ctl(namespace, "stop", id)
	kubelet.Ctl(namespace, "rm", id)
}

func TestPodCommunication(t *testing.T) {
	//should start etcd and flannld
	namespace := "testpod"
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
	success := CreatePod(pod1)
	if !success {
		t.Fatalf("create pod1 failed")
	}
	success = CreatePod(pod2)
	if !success {
		t.Fatalf("create pod2 failed")
	}
	ip1, err := kubelet.GetInfo(namespace, "pod1-c", ".NetworkSettings.IPAddress")
	if err != nil {
		t.Fatalf("get pod1 ip failed")
	}
	ip2, err := kubelet.GetInfo(namespace, "pod2-c", ".NetworkSettings.IPAddress")
	if err != nil {
		t.Fatalf("get pod2 ip failed")
	}
	success = kubelet.CheckCmd(namespace, "pod1-c", []string{"curl", fmt.Sprintf("%s:%s", ip2, pod2.Spec.Containers[0].Env[0].Value)}, "http connect success")
	if !success {
		t.Fatalf("test outside network failed")
	}
	success = kubelet.CheckCmd(namespace, "pod2-c", []string{"curl", fmt.Sprintf("%s:%s", ip1, pod1.Spec.Containers[0].Env[0].Value)}, "http connect success")
	if !success {
		t.Fatalf("test outside network failed")
	}
	//
	client, _ := container.NewClient(namespace)
	ctx := context.Background()
	containers, _ := client.Containers(ctx)
	for _, c := range containers {
		id := c.ID()
		kubelet.Ctl(namespace, "stop", id)
		kubelet.Ctl(namespace, "rm", id)
	}
}
