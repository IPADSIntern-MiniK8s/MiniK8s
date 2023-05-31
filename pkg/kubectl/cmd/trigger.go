package cmd

import (
	"fmt"
	ctlutils "minik8s/pkg/kubectl/utils"
	"minik8s/utils"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var TriggerCmd = &cobra.Command{
	Use:  "trigger <resource> <name> -f <FILENAME>",
	Short: "Kubectl trigger command",
	Long: "Kubectl trigger command, Usage: kubectl trigger <resource> <name> (-f FILENAME)",
	Run: trigger,
}


func trigger(cmd *cobra.Command, args []string) {
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

	kind := strings.ToLower(args[0])
	if kind != "function" && kind != "workflow" {
		fmt.Println("invalid resource type, it should be function or workflow")
	}

	name := strings.ToLower(args[1])
	_url := ctlutils.ParseUrlTrigger(kind, name)
	fmt.Printf("url:%s\n", _url)
	// fmt.Printf("json:%s\n", _json)
	info, err := utils.SendRequest("POST", _json, _url)
	fmt.Println("the response: ", info)
}