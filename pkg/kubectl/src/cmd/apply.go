package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var ApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Kubectl apply manages applications through files defining Kubernetes resources.",
	Long:  "Kubectl apply manages applications through files defining Kubernetes resources. Usage: kubectl apply (-f FILENAME)",
	Run:   apply,
}

func apply(cmd *cobra.Command, args []string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%s", string(content))

}
