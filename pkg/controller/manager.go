package controller

import "minik8s/utils"

func Run() {
	/* service controller */
	var sp svcPodHandler
	var ss svcServiceHandler
	go utils.Sync(ss)
	go utils.Sync(sp)

	/* replicaset controller */
	var rp rsPodHandler
	var rr rsReplicaHandler
	go utils.Sync(rp)
	go utils.Sync(rr)

	utils.WaitForever()
}
