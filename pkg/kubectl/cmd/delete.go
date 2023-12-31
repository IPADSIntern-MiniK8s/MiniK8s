package cmd

import (
	"fmt"
	"log"
	ctlutils "minik8s/pkg/kubectl/utils"
	"minik8s/utils"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wxnacy/wgo/arrays"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete <resource> <name>",
	Short: "Delete resources by resources and names",
	Long:  "Delete resources by resources and names",
	Args:  cobra.ExactArgs(2),
	Run:   delete,
}

func delete(cmd *cobra.Command, args []string) {

	var _url string
	/* get all resources of in certain type under specified namespace */
	kind := strings.ToLower(args[0])
	name := strings.ToLower(args[1])
	/* validate if `kind` is in the resource list */
	if idx := arrays.ContainsString(ctlutils.Resources, kind); idx != -1 {
		_url = ctlutils.ParseUrlOne(kind, name, nameSpace)
	} else if idx := arrays.ContainsString(ctlutils.Globals, kind); idx != -1 {
		_url = ctlutils.ParseUrlOne(kind, name, "nil")
	} else {
		fmt.Printf("error: the server doesn't have a resource type \"%s\"", kind)
	}

	fmt.Printf("url:%s\n", _url)

	/* display the return info */
	var str []byte
	_, err := utils.SendRequest("DELETE", str, _url)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(name, " deleted", "\n")
	/* TODO 解析info，错误判断pod名字是否存在 */
}
