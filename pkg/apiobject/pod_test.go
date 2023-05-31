package apiobject

import (
	"encoding/json"
	"testing"
)

func TestPod(t *testing.T) {
	p := &Pod{
		Data: MetaData{
			Name: "test-pod",
		},
		Spec: PodSpec{
			NodeSelector: map[string]string{"env": "test"},
			Containers: []Container{
				{Name: "test-container"},
			},
		},
	}


	workflowJson, err := json.MarshalIndent(p, "", "    ")
	if err != nil {
		t.Errorf("GenerateWorkflow failed, error marshalling replicas: %s", err)
	}
	t.Log(string(workflowJson))
	
	//expected := `{"metadata":{"name":"test-pod","labels":{}},"spec":{"containers":[{"name":"test-container"}]},"status":{}}`

	// _, err := p.MarshalJSON()
	// if err != nil {
	// 	t.Fatalf("unexpected error: %v", err)
	// }

}
