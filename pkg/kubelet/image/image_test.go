package image

import (
	"minik8s/pkg/kubelet/utils"
	"testing"
)

func TestEnsureImage(t *testing.T) {
	testcases := []string{
		"ubuntu",
		"gpu-server",
		"gpu-server:latest",
		"master:5000/gpu-server",
		"master:5000/gpu-server:latest",
	}
	ns := "test-image"
	client, err := utils.NewClient(ns)
	if err != nil {
		t.Fatalf("client failed")
	}
	for _, i := range testcases {
		utils.Ctl(ns, "rmi", i)
		image := EnsureImage(ns, client, i)
		if image == nil {
			t.Fatalf("pull image %v failed", i)
		}
		image = EnsureImage(ns, client, i)
		if image == nil {
			t.Fatalf("pull exist image %v failed", i)
		}
	}
}
