package apiobject

import (
	"encoding/json"
	"testing"
)

func TestFunction(t *testing.T) {
	function := Function{
		Kind:       "Function",
		APIVersion: "app/v1",
		Name:       "test",
		Path:       "/home/mini-k8s/example/serverless/singlefunc.py",
	}

	functionJson, err := json.MarshalIndent(function, "", "    ")
	if err != nil {
		t.Errorf("GenerateReplicaSet failed, error marshalling replicas: %s", err)
	}

	t.Logf("replicaSet: %s", functionJson)
}
