package cmd

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"log"
	"minik8s/pkg/kubectl/utils"
	"minik8s/utils"
	"os"
	"strings"
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

	kind := strings.ToLower(gjson.Get(string(_json), "kind").String())
	namespace := gjson.Get(string(_json), "metadata.namespace").String()
	_url := ctlutils.ParseUrlMany(kind, namespace)
	fmt.Printf("url:%s\n", _url)

	//utils.SendJsonObject("POST", _json, _url)
	_, err = utils.SendRequest("POST", _json, _url)
	if err != nil {
		log.Fatal(err)
	}
	name := gjson.Get(string(_json), "metadata.name")
	fmt.Print(name, " configured", "\n")

}
