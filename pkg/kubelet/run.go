package kubelet

import (
	"fmt"
	"time"
)

type Config struct {
	ApiserverAddr string // 192.168.1.13:8080
	FlannelSubnet string //10.2.17.1/24
	IP            string //192.168.1.12
	Labels        map[string]string
	ListenAddr    string //localhost:10250
}

func Run(config Config) {
	kl := NewKubelet(config)
	kl.register()
	time.Sleep(time.Second * 5)
	go kl.watchPod()
	err := kl.Server.Run(kl.ListenAddr)
	if err != nil {
		fmt.Println(err.Error())
	}
}
