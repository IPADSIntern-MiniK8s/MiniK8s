package handlers

import "minik8s/pkg/kubeapiserver/storage"

var endpointStorageTool *storage.EtcdStorage = storage.NewEtcdStorageNoParam()
