package controller

import (
	"minik8s/utils"
	"testing"
)

func TestSvcController(t *testing.T) {
	utils.ApiServerIp = "localhost:8080"
}
