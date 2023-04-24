package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	/* flag options that can be inherited by child commands */
	RootCmd.Flags().StringVarP(&nameSpace, "nameSpace", "n", "default", "kubectl (-n NAMESPACE)")

	/* apply cmd: eg: kubectl apply -f <FILENAME> */
	ApplyCmd.Flags().StringVarP(&filePath, "filePath", "f", "", "kubectl apply -f <FILENAME>")
	ApplyCmd.MarkFlagRequired("filePath")
	RootCmd.AddCommand(ApplyCmd)

	RootCmd.AddCommand(GetCmd)

}

var filePath string
var nameSpace string

var RootCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "kubectl controls the minik8s cluster manager.",
	Long:  "kubectl controls the minik8s cluster manager.",
	Run:   runRoot,
}

func runRoot(cmd *cobra.Command, args []string) {
	fmt.Printf("execute %s args:%v \n", cmd.Name(), args)
}