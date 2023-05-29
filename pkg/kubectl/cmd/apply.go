package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	ctlutils "minik8s/pkg/kubectl/utils"
	"minik8s/utils"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
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
		return
	}

	_json, err := yaml.YAMLToJSON(_yaml)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	kind := strings.ToLower(gjson.Get(string(_json), "kind").String())
	namespace := gjson.Get(string(_json), "metadata.namespace").String()
	if namespace == "" {
		var obj map[string]interface{}
		json.Unmarshal(_json, &obj)
		obj["metadata"].(map[string]interface{})["namespace"] = "default"
		_json, _ = json.Marshal(obj)
	}
	_url := ctlutils.ParseUrlMany(kind, namespace)
	fmt.Printf("url:%s\n", _url)
	_, err = utils.SendRequest("POST", _json, _url)
	if err != nil {
		log.Fatal(err)
	}
	name := gjson.Get(string(_json), "metadata.name")
	fmt.Print(name, " configured", "\n")

}
