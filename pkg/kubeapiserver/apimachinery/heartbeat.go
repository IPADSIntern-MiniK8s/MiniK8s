package apimachinery

import (
	"context"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/apiobject/utils"
	"minik8s/pkg/kubeapiserver/storage"
	"time"
	log "github.com/sirupsen/logrus"
)

func HeartBeat() {
	// get all node from etcd
	storageTool := storage.NewEtcdStorageNoParam()
	key := "/registry/nodes/"
	for {
		log.Info("[HeartBeat] check node check heartbeat")
		var nodes []apiobject.Node
		err := storageTool.GetList(context.Background(), key, &nodes)
		if err != nil {
			log.Error("[HeartBeat] the node list is empty")
		} else {
			for _, node := range nodes {
				// check timeout
				if node.Status.Time == 0 {  // the time not assigned
					continue
				}
				timeout := utils.CheckTimeout(node.Status.Time)
				if timeout {
					nodeKey := "/registry/nodes/" + node.Data.Name
					node.Status.Conditions[0].Status = apiobject.NetworkUnavailable
					err := storageTool.GuaranteedUpdate(context.Background(), nodeKey, &node)
					if err != nil {
						log.Info("[HeartBeat] update node error: ", err)
					}
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}