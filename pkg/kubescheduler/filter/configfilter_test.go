package filter

import (
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubescheduler/testutils"
	"testing"
)

func TestConfigFilter_PreFilter(t *testing.T) {
	var pod *apiobject.Pod
	filter := NewConfigFilter()

	// test empty pod
	ret := filter.PreFilter(pod)
	if ret != false {
		t.Error("[TestConfigFilter_PreFilter] test empty pod fail")
	}

	// test illegal pod
	pod = testutils.CreateIllegalPod()
	ret = filter.PreFilter(pod)
	if ret != false {
		t.Error("[TestConfigFilter_PreFilter] test illegal pod fail")
	}

	// test legal pod
	pod = testutils.CreatePod()
	ret = filter.PreFilter(pod)
	if ret != true {
		t.Error("[TestConfigFilter_PreFilter] test legal pod fail")
	}
}

func TestConfigFilter_CheckNodeStatus(t *testing.T) {
	nodes := make([]*apiobject.Node, 0)
	node1 := testutils.CreateNode("test-node1", apiobject.Ready, "100m", "256Mi", "200m", "512Mi", "192.168.1.13")
	nodes = append(nodes, node1)

	node2 := testutils.CreateNode("test-node2", apiobject.NetworkUnavailable, "100m", "256Mi", "200m", "512Mi", "192.168.1.13")
	nodes = append(nodes, node2)

	node3 := testutils.CreateNode("test-node3", apiobject.Ready, "100m", "256Mi", "200m", "512Mi", "")
	nodes = append(nodes, node3)

	filter := NewConfigFilter()

	result := filter.CheckNodeStatus(nodes)
	if len(result) != 1 {
		t.Error("[TestConfigFilter_CheckNodeStatus] test fail")
	}
}

func TestConfigFilter_CheckNodeSelector(t *testing.T) {
	nodes := make([]*apiobject.Node, 0)
	node1 := testutils.CreateNode("test-node1", apiobject.Ready, "100m", "256Mi", "200m", "512Mi", "192.168.1.13")
	nodes = append(nodes, node1)

	node2 := testutils.CreateNode("test-node2", apiobject.Ready, "100m", "256Mi", "200m", "512Mi", "192.168.1.14")
	node2.Data.Labels = map[string]string{
		"disktype": "ssd",
	}
	nodes = append(nodes, node2)

	filter := NewConfigFilter()

	pod := testutils.CreatePod()

	result := filter.CheckNodeSelector(pod, nodes)

	if len(result) != 1 {
		t.Error("[TestConfigFilter_CheckNodeSelector] test fail")
	}
}

func TestConfigFilter_GetResourceRequest(t *testing.T) {
	pod := testutils.CreatePod()
	filter := NewConfigFilter()

	cpu, memory := filter.GetResourceRequest(pod)

	log.Info("[TestConfigFilter_GetResourceRequest] cpu: ", cpu, " memory: ", memory)
}

func TestConfigFilter_CheckResource(t *testing.T) {
	nodes := make([]*apiobject.Node, 0)
	node1 := testutils.CreateNode("test-node1", apiobject.Ready, "100m", "256Mi", "200m", "512Mi", "192.168.119.132")
	nodes = append(nodes, node1)

	node2 := testutils.CreateNode("test-node2", apiobject.Ready, "150m", "256Mi", "300m", "345Mi", "192.168.119.128")
	nodes = append(nodes, node2)

	node3 := testutils.CreateNode("test-node3", apiobject.Ready, "10m", "10Mi", "20m", "20Mi", "192.168.119.134")
	nodes = append(nodes, node3)

	filter := NewConfigFilter()
	pod := testutils.CreatePod()

	cpuRequest, memoryRequest := filter.GetResourceRequest(pod)

	result := filter.CheckResource(cpuRequest, memoryRequest, nodes)
	if len(result) != 2 {
		t.Error("[TestConfigFilter_CheckResource] test fail")
	}
}
