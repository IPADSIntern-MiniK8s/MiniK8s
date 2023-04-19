package handlers

import (
	"github.com/coreos/etcd/clientv3"
	"minik8s/pkg/kubeapiserver/storage"
)

type HandlerStorage struct {
	StorageTool *storage.EtcdStorage
}

func NewHandlerStorage() *HandlerStorage {
	var client, _ = clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2380"},
	})
	return &HandlerStorage{
		StorageTool: storage.NewEtcdStorage(client),
	}
}
