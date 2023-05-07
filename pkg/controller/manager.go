package controller

import "minik8s/utils"

func run() {
	var sp svcPodHandler
	var ss svcServiceHandler
	go utils.Sync(ss)
	go utils.Sync(sp)
	utils.WaitForever()
}
