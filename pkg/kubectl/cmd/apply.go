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
	"github.com/wxnacy/wgo/arrays"
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
	if kind == "dnsrecord" {
		namespace = gjson.Get(string(_json), "namespace").String()
	}
	var _url string
	//fmt.Print(namespace)
	if idx := arrays.ContainsString(ctlutils.Resources, kind); idx != -1 {
		if namespace == "" {
			var obj map[string]interface{}
			json.Unmarshal(_json, &obj)
			obj["metadata"].(map[string]interface{})["namespace"] = "default"
			_json, _ = json.Marshal(obj)
		}
		_url = ctlutils.ParseUrlMany(kind, namespace)
	} else if idx := arrays.ContainsString(ctlutils.Globals, kind); idx != -1 {
		_url = ctlutils.ParseUrlMany(kind, "nil")
	} else {
		fmt.Printf("error: the server doesn't have a resource type \"%s\"", kind)
	}
	fmt.Printf("url:%s\n", _url)
	info, err := utils.SendRequest("POST", _json, _url)
	if err != nil {
		log.Fatal(info)
	}
	name := gjson.Get(string(_json), "metadata.name")
	fmt.Print(name, " configured", "\n")

}
