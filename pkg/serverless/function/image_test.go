package function

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCreateImage(t *testing.T) {
	err := CreateImage("/home/mini-k8s/example/serverless/singlefunc.py", "test")
	if err != nil {
		t.Errorf("CreateImage failed, error: %s", err)
	}
}

func TestDeleteImage(t *testing.T) {
	err := DeleteImage("test")
	if err != nil {
		t.Errorf("DeleteImage failed, error: %s", err)
	}

	// search the image
	cmd := exec.Command("docker", "images")
	out, err := cmd.Output()
	if err != nil {
		t.Errorf("DeleteImage failed, error: %s", err)
	}
	
	outputStr := string(out)
	imageName := "localhost:5000/test:latest"
	if strings.Contains(outputStr, imageName) {
		t.Errorf("DeleteImage failed, image %s still exists", imageName)
	}
}


