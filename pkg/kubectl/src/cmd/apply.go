package cmd

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	ctlutils "minik8s/pkg/kubectl/src/utils"
	"minik8s/utils"
	"os"
)

var ApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Kubectl apply manages applications through files defining Kubernetes resources.",
	Long:  "Kubectl apply manages applications through files defining Kubernetes resources. Usage: kubectl apply (-f FILENAME)",
	Run:   apply,
}

func apply(cmd *cobra.Command, args []string) {
	_yaml, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
	}

	_json, err := yaml.YAMLToJSON(_yaml)
	if err != nil {
		fmt.Println(err.Error())
	}

	_url := ctlutils.ParseUrlFromJson(_json)
	fmt.Printf("url:%s\n", _url)

	utils.SendJsonObject("POST", _json, _url)
}
