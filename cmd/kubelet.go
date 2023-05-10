package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/pkg/kubelet"
	"minik8s/utils"
)

var RootCmd = &cobra.Command{
	Use:   "kubelet",
	Short: "kubelet manages containers",
	Long:  "kubelet manages containers",
	Run:   runRoot,
}

var apiserverAddr string

func init() {
	RootCmd.Flags().StringVarP(&apiserverAddr, "apiserver-address", "a", utils.ApiServerIp, "kubelet (-a apiserver-address)")
}
func runRoot(cmd *cobra.Command, args []string) {
	kubelet.Run(apiserverAddr)
}
func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
	}
}
