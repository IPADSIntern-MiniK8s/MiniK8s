package kubelet

import (
	"fmt"
	"minik8s/config"
	"time"
)

type Config struct {
	ApiserverAddr string // 192.168.1.13:8080
	FlannelSubnet string //10.2.17.1/24
	IP            string //192.168.1.12
	Labels        map[string]string
	ListenAddr    string //localhost:10250
	CPU           string
	Memory        string
}

func Run(c Config) {
	config.ApiServerIp = c.ApiserverAddr
	kl := NewKubelet(c)
	go func() {
		for {
			kl.register()
			time.Sleep(time.Second * 5)
			//normally, watch Pod will not return
			kl.watchPod()
			fmt.Println("trying to reconnect to apiserver")
			time.Sleep(time.Second * 5)
		}
	}()

	go kl.watchContainersStatus()
	err := kl.Server.Run(kl.ListenAddr)
	if err != nil {
		fmt.Println(err.Error())
	}
}
