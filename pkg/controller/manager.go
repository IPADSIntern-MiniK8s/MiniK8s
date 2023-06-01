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

	/* hpa controller */
	var hs hpaScalerHandler
	go utils.Sync(hs)
	go regularCheck()

	var jc jobPodHandler
	var jh jobHandler
	go utils.Sync(jc)
	go utils.Sync(jh)


	utils.WaitForever()
}
