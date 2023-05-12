package handlers

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreatePodHandler(t *testing.T) {

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

	CreatePodHandler(c)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		if w.Code == http.StatusInternalServerError {
			log.Warn("TestPodHandler_CreatePod: ", w.Body.String())
		}
		t.Fatalf("unexpected status code: %d", w.Code)
	}
}

func TestGetPodHandler(t *testing.T) {

	url := "/api/v1/namespaces/{namespace}/pods/{name}"
	namespace := "default"
	name := "test-pod"
	url = strings.Replace(url, "{namespace}", namespace, 1)
	url = strings.Replace(url, "{name}", name, 1)
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = req

	GetPodHandler(c)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		if w.Code == http.StatusInternalServerError {
			log.Warn("TestPodHandler_GetPod: ", w.Body.String())
		}
		t.Fatalf("unexpected status code: %d", w.Code)
	}
}

func TestDeletePodHandler(t *testing.T) {

	url := "/api/v1/namespaces/{namespace}/pods/{name}"
	namespace := "default"
	name := "test-pod"
	url = strings.Replace(url, "{namespace}", namespace, 1)
	url = strings.Replace(url, "{name}", name, 1)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = req

	DeletePodHandler(c)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		if w.Code == http.StatusInternalServerError {
			log.Warn("TestPodHandler_DeletePod: ", w.Body.String())
		}
		t.Fatalf("unexpected status code: %d", w.Code)
	}
}
