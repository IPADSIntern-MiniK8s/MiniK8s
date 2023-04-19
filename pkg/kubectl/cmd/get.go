package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wxnacy/wgo/arrays"
	"log"
	ctlutils "minik8s/pkg/kubectl/src/utils"
	"minik8s/utils"
	"strings"
)

var GetCmd = &cobra.Command{
	Use:   "get <resource> <name>",
	Short: "Display one or many resources",
	Long:  "Display one or many resources",
	Args:  cobra.RangeArgs(1, 2),
	Run:   get,
}

func get(cmd *cobra.Command, args []string) {

	var _url string
	if len(args) == 1 {
		/* get all resources of in certain type under specified namespace */
		kind := strings.ToLower(args[0])
		kind = kind[0 : len(kind)-1]
		/* validate if `kind` is in the resource list */
		if idx := arrays.ContainsString(ctlutils.Resources, kind); idx == -1 {
			fmt.Printf("error: the server doesn't have a resource type \"%s\"", kind)
		}

		_url = ctlutils.ParseUrlMany(kind, nameSpace)
		fmt.Printf("url:%s\n", _url)

	} else {
		/* get resource in certain type with its name under specified namespace */
		kind := strings.ToLower(args[0])
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
	info, err := utils.SendRequest("GET", str, _url)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(info)
	/* TODO 解析info，错误判断pod名字是否存在 */
}
