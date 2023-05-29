package main

import (
	"fmt"
	"minik8s/config"
	"minik8s/pkg/kubeproxy"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ProxyCmd = &cobra.Command{
	Use:   "kubeproxy",
	Short: "kubeproxy manages network",
	Long:  "kubeproxy manages network",
	Run:   runProxy,
}

var proxyConfigAddr string

var KubeproxyConfig = Config{
	ApiserverAddr: "192.168.1.13:8080",
}

type Config struct {
	ApiserverAddr string // 192.168.1.13:8080
}

func proxyInitConfig() {
	//fmt.Println(configAddr)
	viper.SetConfigFile(proxyConfigAddr)
	err := viper.ReadInConfig()
	if err == nil {
		//panic(err)
		if err := viper.Unmarshal(&KubeproxyConfig); err != nil {
			//panic(err)
		}
	}
	//if err,use default config
	fmt.Println(KubeproxyConfig)
}

func init() {
	cobra.OnInitialize(proxyInitConfig)
	//RootCmd.Flags().StringVarP(&apiserverAddr, "apiserver-address", "a", utils.ApiServerIp, "kubelet (-a apiserver-address)")
	ProxyCmd.PersistentFlags().StringVarP(&proxyConfigAddr, "config", "c", "./kubeproxy-config.yaml", "kubeproxy (-c config)")
}

func runProxy(cmd *cobra.Command, args []string) {
	config.ApiServerIp = KubeproxyConfig.ApiserverAddr
	kubeproxy.Run()
}

func main() {
	if err := ProxyCmd.Execute(); err != nil {
		fmt.Println(err.Error())
	}
}
