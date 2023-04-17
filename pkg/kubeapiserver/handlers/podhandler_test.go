package handlers

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/gin-gonic/gin"
	"minik8s/pkg/apiobject"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPodHandler_CreatePod(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2380"},
	})

	if err != nil {
		t.Fatal(err)
	}

	podHandler := NewPodHandler(client)
	p := &apiobject.Pod{
		Data: apiobject.MetaData{
			Name: "test-pod",
		},
		Spec: apiobject.PodSpec{
			Containers: []apiobject.Container{
				{Name: "test-container"},
			},
		},
	}

	requestBody, err := p.MarshalJSON()
	payload := strings.NewReader(string(requestBody))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	url := "/api/v1/namespaces/{namespace}/pods"
	namespace := "default"
	url = strings.Replace(url, "{namespace}", namespace, 1)
	req, _ := http.NewRequest(http.MethodPost, url, payload)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = req

	podHandler.CreatePod(c)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", w.Code)
	}
}
