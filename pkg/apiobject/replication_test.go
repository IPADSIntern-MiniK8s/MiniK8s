package apiobject

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestReplicationController(t *testing.T) {
	replica := ReplicationController{
		APIVersion: "v1",
		Data: MetaData{
			Name:      "deploy-practice",
			Namespace: "default",
		},
		Spec: ReplicationControllerSpec{
			Replicas: 3,
			Selector: map[string]string{
				"app": "deploy-practice",
			},
		},
	}
	jsonBytes, _ := json.MarshalIndent(replica, "", "    ")
	fmt.Println(string(jsonBytes))

}
