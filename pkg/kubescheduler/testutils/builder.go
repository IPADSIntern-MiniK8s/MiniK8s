package testutils

import "minik8s/pkg/apiobject"

func CreatePod() *apiobject.Pod {
	return &apiobject.Pod{
		APIVersion: "v1",
		Data: apiobject.MetaData{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: apiobject.PodSpec{
			NodeSelector: map[string]string{
				"disktype": "ssd",
			},
			Containers: []apiobject.Container{
				{
					Name:  "test-container",
					Image: "nginx",
					Resources: apiobject.Resources{
						Limits: apiobject.Limit{
							Cpu:    "200m",
							Memory: "512Mi",
						},
						Requests: apiobject.Request{
							Cpu:    "100m",
							Memory: "256Mi",
						},
					},
				},
			},
		},
		Status: apiobject.PodStatus{
			Phase:  apiobject.Pending,
			HostIp: "",
			PodIp:  "",
		},
	}
}

func CreateIllegalPod() *apiobject.Pod {
	return &apiobject.Pod{
		APIVersion: "v1",
		Data: apiobject.MetaData{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: apiobject.PodSpec{
			NodeSelector: map[string]string{
				"disktype": "ssd",
			},
			Containers: []apiobject.Container{
				{
					Name:  "test-container",
					Image: "nginx",
					Resources: apiobject.Resources{
						Limits: apiobject.Limit{
							Cpu:    "200m",
							Memory: "512Mi",
						},
						Requests: apiobject.Request{
							Cpu:    "100m",
							Memory: "256Mi",
						},
					},
				},
			},
		},
		Status: apiobject.PodStatus{
			Phase:  apiobject.Running,
			HostIp: "",
			PodIp:  "",
		},
	}
}

// create node for test use
func CreateNode(name string, status apiobject.NodeStatusTag, cpuMin string, memoryMin string, cpuMax string, memoryMax string, ip string) *apiobject.Node {
	return &apiobject.Node{
		APIVersion: "v1",
		Data: apiobject.MetaData{
			Name:      name,
			Namespace: "default",
		},
		Spec: apiobject.NodeSpec{
			Unschedulable: false,
			PodCIDR:       "10.100.10.14/24",
		},
		Status: apiobject.NodeStatus{
			Capability: map[string]string{
				"cpu":    cpuMax,
				"memory": memoryMax,
			},
			Allocatable: map[string]string{
				"cpu":    cpuMin,
				"memory": memoryMin,
			},
			Conditions: []apiobject.Condition{
				{
					Status: status,
				},
			},
			Addresses: []apiobject.Address{
				{
					Type:    "InternalIP",
					Address: ip,
				},
			},
		},
	}
}
