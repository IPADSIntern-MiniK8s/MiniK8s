package policy

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubescheduler/filter"
	"minik8s/pkg/kubescheduler/testutils"
	"testing"
)

func TestConfigScheduler_Schedule(t *testing.T) {
	pod := testutils.CreatePod()

	// print the pod
	jsonBytes, _ := json.MarshalIndent(pod, "", "    ")
	fmt.Println(string(jsonBytes))

	nodes := make([]*apiobject.Node, 0)
	node1 := testutils.CreateNode("test-node1", apiobject.Ready, "100m", "256Mi", "200m", "512Mi", "192.168.119.132")
	node1.Data.Labels = map[string]string{
		"disktype": "ssd",
	}

	nodes = append(nodes, node1)

	node2 := testutils.CreateNode("test-node2", apiobject.Ready, "150m", "256Mi", "300m", "345Mi", "192.168.119.128")
	node2.Data.Labels = map[string]string{
		"disktype": "ssd",
	}
	nodes = append(nodes, node2)

	node3 := testutils.CreateNode("test-node3", apiobject.Ready, "10m", "10Mi", "20m", "20Mi", "192.168.119.134")
	nodes = append(nodes, node3)

	node4 := testutils.CreateNode("test-node4", apiobject.NetworkUnavailable, "100m", "256Mi", "200m", "512Mi", "192.168.1.13")
	nodes = append(nodes, node4)

	for _, node := range nodes {
		jsonBytes, _ := json.MarshalIndent(node, "", "    ")
		fmt.Println(string(jsonBytes))
	}

	concreteFilter := filter.NewConfigFilter()
	var f filter.TemplateFilter
	f = concreteFilter
	scheduler := NewResourceScheduler(f)
	selectedNodes := scheduler.Schedule(pod, nodes)
	if len(selectedNodes) != 2 {
		t.Errorf("expected 2 nodes, but got %d", len(selectedNodes))
	}

	for _, node := range selectedNodes {
		log.Info("[TestConfigScheduler_Schedule] selected node: ", node.Data.Name)
	}

}
