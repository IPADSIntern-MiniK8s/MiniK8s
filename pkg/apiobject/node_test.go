package apiobject

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestNode(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	n := &Node{
		Kind:       "Node",
		APIVersion: "v1",
		Data: MetaData{
			Name: "test",
		},
		Spec: NodeSpec{
			Unschedulable: false,
		},
	}

	b, err := n.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("Node string: ", string(b))

	n2 := &Node{}
	err = n2.UnMarshalJSON(b)
	if err != nil {
		t.Fatal(err)
	}
}
