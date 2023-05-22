package activator

import (
	"encoding/json"
	"testing"
)

func generateImage(name string) string {
	return serverIp + ":5000/" + name + ":latest"
}

func TestGenerateReplicaSet(t *testing.T) {
	replica := GenerateReplicaSet("test", "serverless", generateImage("test"), 0)
	if replica.Data.Name != "test" {
		t.Errorf("GenerateReplicaSet failed, expected %s, got %s", "test", replica.Data.Name)
	}
	if replica.Data.Namespace != "serverless" {
		t.Errorf("GenerateReplicaSet failed, expected %s, got %s", "serverless", replica.Data.Namespace)
	}
	if replica.Spec.Replicas != 0 {
		t.Errorf("GenerateReplicaSet failed, expected %d, got %d", 0, replica.Spec.Replicas)
	}

	// print the replicaSet
	replicaJson, err := json.MarshalIndent(replica, "", "    ")
	if err != nil {
		t.Errorf("GenerateReplicaSet failed, error marshalling replicas: %s", err)
	}

	t.Logf("replicaSet: %s", replicaJson)
}



