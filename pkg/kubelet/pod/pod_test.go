package pod

import (
	"context"
	"fmt"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubelet"
	"minik8s/pkg/kubelet/container"
	"testing"
)

func TestParseMemory(t *testing.T) {
	res, err := parseMemory("100Mi")
	if err != nil || res != 100*1024*1024 {
		t.Fatalf("test parseMemory error")
	}
	res, err = parseMemory("10000")
	if err != nil || res != 10000 {
		t.Fatalf("test parseMemory failed")
	}
}

func TestParseCmd(t *testing.T) {
	res := parseCmd([]string{"/bin/bash"}, []string{"-c", "echo Hello Kubernetes!"})
	if len(res) != 3 || res[0] != "/bin/bash" || res[1] != "-c" || res[2] != "echo Hello Kubernetes!" {
		t.Fatalf("test parsecmd failed")
	}
}

func TestParseEnv(t *testing.T) {
	res := parseEnv([]apiobject.Env{
		{"a", "b"}, {"c", "d"},
	})
	if len(res) != 2 || res[0] != "a=b" || res[1] != "c=d" {
		t.Fatalf("test parseenv failed")
	}
}

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
	success = kubelet.CheckCmd(namespace, "testpod-c2", []string{"ping", "www.baidu.com", "-c", "2"}, "64 bytes from")
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
