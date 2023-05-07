package handlers

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubeapiserver/watch"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterNodeHandler(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	n := &apiobject.Node{
		Kind: "Node",
		Data: apiobject.MetaData{
			Name: "test-node",
		},
		Spec: apiobject.NodeSpec{
			Unschedulable: false,
		},
	}

	requestBody, err := n.MarshalJSON()
	log.Debug("TestRegisterNodeHandler the request body is ", string(requestBody), " and the error is ", err)
	if err != nil {
		t.Error(err)
	}

	payload := strings.NewReader(string(requestBody))
	url := "api/v1/nodes"
	req, err := http.NewRequest(http.MethodGet, url, payload)
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = req

	RegisterNodeHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		log.Debug("[TestRegisterNodeHandler] ", w.Body.String())
	}

	watcherKey := "registry/nodes/test"
	_, OK := watch.WatchTable[watcherKey]
	if !OK {
		t.Errorf("Expected watcher key %s, got %s", watcherKey, "nil")
	}

}
