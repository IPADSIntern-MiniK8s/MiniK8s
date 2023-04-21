package config

import "minik8s/pkg/kubeapiserver/apimachinery"

// WatchTable map the attribute name to the watch server
var WatchTable map[string]*apimachinery.WatchServer
