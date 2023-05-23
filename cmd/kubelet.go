package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"minik8s/pkg/kubelet"
)

var RootCmd = &cobra.Command{
	Use:   "kubelet",
	Short: "kubelet manages containers",
	Long:  "kubelet manages containers",
	Run:   runRoot,
}

var configAddr string

//	var KubeletConfig = kubelet.Config{
//		ApiserverAddr: "192.168.1.13:8080",
//		FlannelSubnet: "10.2.17.1/24",
//		IP:            "192.168.1.12",
//		Labels:        map[string]string{},
//		ListenAddr:    "localhost:10250",
//	}
var KubeletConfig = kubelet.Config{
	ApiserverAddr: "192.168.1.14:8080",
	FlannelSubnet: "10.2.9.1/24",
	IP:            "192.168.1.14",
	Labels:        map[string]string{},
	ListenAddr:    "localhost:10250",
}

func initConfig() {
	//fmt.Println(configAddr)
	viper.SetConfigFile(configAddr)
	err := viper.ReadInConfig()
	if err == nil {
		//panic(err)
		if err := viper.Unmarshal(&KubeletConfig); err != nil {
			//panic(err)
		}
	}
	//if err,use default config
	fmt.Println(KubeletConfig)
}

func init() {
	cobra.OnInitialize(initConfig)
	//RootCmd.Flags().StringVarP(&apiserverAddr, "apiserver-address", "a", utils.ApiServerIp, "kubelet (-a apiserver-address)")
	RootCmd.PersistentFlags().StringVarP(&configAddr, "config", "c", "./kubelet-config.yaml", "kubelet (-c config)")
}

func runRoot(cmd *cobra.Command, args []string) {
	kubelet.Run(KubeletConfig)
}
func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
	}
}
