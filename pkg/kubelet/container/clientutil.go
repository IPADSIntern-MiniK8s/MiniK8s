package container

import (
	"github.com/containerd/containerd"
)

func NewClient() (*containerd.Client, error) {
	return containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace("minik8s"))
}
