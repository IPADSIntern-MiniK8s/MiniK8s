package container

import (
	"github.com/containerd/containerd"
)

func NewClient(namespace string) (*containerd.Client, error) {
	return containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
}
