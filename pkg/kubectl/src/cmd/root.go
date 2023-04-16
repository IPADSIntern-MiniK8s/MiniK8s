package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	ApplyCmd.Flags().StringVarP(&filePath, "filePath", "f", "", "kubectl apply (-f FILENAME)")
	RootCmd.AddCommand(ApplyCmd)
}

var filePath string

var RootCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "Kubectl is for interactive control of minik8s",
	Long:  `Kubectl is for interactive control of minik8s`,
	Run:   runRoot,
}

func runRoot(cmd *cobra.Command, args []string) {
	fmt.Printf("execute %s args:%v \n", cmd.Name(), args)
}
