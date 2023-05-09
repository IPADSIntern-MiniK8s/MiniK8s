package policy

import (
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubescheduler/filter"
	"minik8s/pkg/kubescheduler/testutils"
	"testing"
)

func TestResourceScheduler_Schedule(t *testing.T) {
	pod := testutils.CreatePod()

	// create nodes
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

	node3 := testutils.CreateNode("test-node3", apiobject.Ready, "150m", "256Mi", "300m", "345Mi", "192.168.119.128")
	node3.Data.Labels = map[string]string{
		"disktype": "ssd",
	}
	nodes = append(nodes, node3)

	// create scheduler
	concreteFilter := filter.NewConfigFilter()
	var f filter.TemplateFilter
	f = concreteFilter
	scheduler := NewLeastRequestScheduler(f)
	// schedule pod for three times
	for i := 0; i < 3; i++ {
		log.Info("[TestResourceScheduler_Schedule] schedule pod for ", i, " time[s]")
		selectedNode := scheduler.Schedule(pod, nodes)
		for _, n := range selectedNode {
			log.Info("[TestResourceScheduler_Schedule] the selected node is: ", n.Data.Name)
		}
	}

}
