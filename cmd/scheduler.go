package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"minik8s/pkg/kubescheduler"
)

var SchedulerCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  "",
	Run:   runScheduler,
}

var schedulerConfigAddr string

var SchedulerConfig = kubescheduler.Config{
	ApiserverAddr: "192.168.1.13:8080",
	Policy:        "default",
}

func schedulerInitConfig() {
	//fmt.Println(configAddr)
	viper.SetConfigFile(schedulerConfigAddr)
	err := viper.ReadInConfig()
	if err == nil {
		//panic(err)
		if err := viper.Unmarshal(&SchedulerConfig); err != nil {
			//panic(err)
		}
	}
	//if err,use default config
	fmt.Println(SchedulerConfig)
}

func init() {
	cobra.OnInitialize(schedulerInitConfig)
	//RootCmd.Flags().StringVarP(&apiserverAddr, "apiserver-address", "a", utils.ApiServerIp, "kubelet (-a apiserver-address)")
	SchedulerCmd.PersistentFlags().StringVarP(&schedulerConfigAddr, "config", "c", "./scheduler-config.yaml", "scheduler (-c config)")
}

func runScheduler(cmd *cobra.Command, args []string) {
	kubescheduler.Run(SchedulerConfig)
}
func main() {
	if err := SchedulerCmd.Execute(); err != nil {
		fmt.Println(err.Error())
	}
}

//func main() {
//	// use parameter to get the policy
//	policy := "default"
//	if len(os.Args) > 1 {
//		policy = os.Args[1]
//	}
//	kubescheduler.Run(policy)
//}
