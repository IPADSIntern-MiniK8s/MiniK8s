package apiobject

import "testing"

func TestPod(t *testing.T) {
	p := &Pod{
		Data: MetaData{
			Name: "test-pod",
		},
		Spec: PodSpec{
			Containers: []Container{
				{Name: "test-container"},
			},
		},
	}

	expected := `{"metadata":{"name":"test-pod","labels":{}},"spec":{"containers":[{"name":"test-container"}]},"status":{}}`

	b, err := p.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(b) != expected {
		t.Errorf("got %s, want %s", string(b), expected)
	}
}
