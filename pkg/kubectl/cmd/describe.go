package cmd

import (
	"fmt"
	ctlutils "minik8s/pkg/kubectl/utils"
	log "github.com/sirupsen/logrus"
	"minik8s/utils"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/wxnacy/wgo/arrays"
)

var DescribeCmd = &cobra.Command{
	Use:   "describe <resource> <name>/ describe <resource>s",
	Short: "Display one or many resources",
	Long:  "Display one or many resources",
	Args:  cobra.RangeArgs(1, 2),
	Run:   describe,
}

func describe(cmd *cobra.Command, args []string) {

	var _url string
	var kind string
	if len(args) == 1 {
		/* get all resources of in certain type under specified namespace */
		kind = strings.ToLower(args[0])
		kind = kind[0 : len(kind)-1]
		/* validate if `kind` is in the resource list */
		if idx := arrays.ContainsString(ctlutils.Resources, kind); idx == -1 {
			fmt.Printf("error: the server doesn't have a resource type \"%s\"", kind)
		}

		_url = ctlutils.ParseUrlMany(kind, nameSpace)
		fmt.Printf("url:%s\n", _url)

	} else {
		/* get resource in certain type with its name under specified namespace */
		kind = strings.ToLower(args[0])
		name := strings.ToLower(args[1])
		/* validate if `kind` is in the resource list */
		if idx := arrays.ContainsString(ctlutils.Resources, kind); idx == -1 {
			fmt.Printf("error: the server doesn't have a resource type \"%s\"", kind)
		}

		_url = ctlutils.ParseUrlOne(kind, name, nameSpace)
		fmt.Printf("url:%s\n", _url)

	} 

	/* display the info */
	var str []byte
	_json, err := utils.SendRequest("GET", str, _url)
	if err != nil {
		//log.Fatal(err)
		/* 解析info，错误判断pod名字是否存在 */
		fmt.Print(_json)
	}


	if kind == "function" {
		path := gjson.Get(string(_json), "path")
		if path.Exists() {
			f, err := os.Open("example.txt")
			if err != nil {
				log.Info("open file error")
			}
			defer f.Close()
			buf := make([]byte, 1024)
			// read the entire file
			n, err := f.Read(buf)
			fmt.Print(string(buf[:n]))
			fmt.Print("\n")
		}
	} else {
		info, _ := yaml.JSONToYAML([]byte(_json))
		fmt.Print(string(info))
		fmt.Print("\n")
	}
	

	/* {"error":"key not found: /registry/pods/default/test"} */
	/* {"metadata":{"name":"test-pod"},"spec":{"containers":[{"name":"test-container","resources":{"limits":{},"requests":{}}}]},"status":{"phase":"Pending"}}root@minik8s-2:~/mini-k8s/pkg/kubectl/test# */

}
