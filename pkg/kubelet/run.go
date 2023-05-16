package kubelet

import (
	"time"
)

type Config struct {
	ApiserverAddr string // 192.168.1.13:8080
	FlannelSubnet string //10.2.17.1/24
	IP            string //192.168.1.12
	Labels        map[string]string
}

func Run(config Config) {
	kl := NewKubelet(config)
	kl.register()
	time.Sleep(time.Second * 5)
	kl.watchPod()
}
