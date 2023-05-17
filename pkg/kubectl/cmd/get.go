package cmd

import (
	"fmt"
	"github.com/liushuochen/gotable"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/wxnacy/wgo/arrays"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/kubectl/utils"
	"minik8s/utils"
	"strconv"
	"strings"
)

var GetCmd = &cobra.Command{
	Use:   "get <resource> <name>/ get <resource>s",
	Short: "Display one or many resources",
	Long:  "Display one or many resources",
	Args:  cobra.RangeArgs(1, 2),
	Run:   get,
}

func get(cmd *cobra.Command, args []string) {

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
	switch kind {
	case "pod":
		{
			table, _ := gotable.Create("NAME", "POD-IP", "STATUS")
			podList := gjson.Parse(_json).Array()
			for _, p := range podList {
				name := gjson.Get(p.String(), "metadata.name").String()
				status := gjson.Get(p.String(), "status.phase").String()
				IP := gjson.Get(p.String(), "status.podIP").String()
				table.AddRow(map[string]string{
					"NAME":   name,
					"POD-IP": IP,
					"STATUS": status,
				})
			}
			fmt.Println(table)
		}
	case "service":
		{
			table, _ := gotable.Create("NAME", "TYPE", "CLUSTER-IP", "PORT(S)")
			svcList := gjson.Parse(_json).Array()
			for _, svc := range svcList {
				name := gjson.Get(svc.String(), "metadata.name").String()
				ty := gjson.Get(svc.String(), "spec.type").String()
				ip := gjson.Get(svc.String(), "status.clusterIP").String()
				ports := gjson.Get(svc.String(), "spec.ports").Array()
				portString := ""
				for _, p := range ports {
					port := gjson.Get(p.String(), "port").Int()
					protocal := gjson.Get(p.String(), "protocol").String()
					portString += fmt.Sprintf("%d/%s,", port, protocal)
				}

				table.AddRow(map[string]string{
					"NAME":       name,
					"TYPE":       ty,
					"CLUSTER-IP": ip,
					"PORTS":      portString,
				})
			}
			fmt.Println(table)
		}
	case "endpoint":
		{
			table, _ := gotable.Create("NAME", "SERVICE-IP", "POD-IP")
			edpList := gjson.Parse(_json).Array()
			for _, p := range edpList {
				edpt := &apiobject.Endpoint{}
				edpt.UnMarshalJSON([]byte(p.String()))
				table.AddRow(map[string]string{
					"NAME":       edpt.Data.Name,
					"POD-IP":     edpt.Spec.DestIP + ":" + strconv.Itoa(int(edpt.Spec.DestPort)),
					"SERVICE-IP": edpt.Spec.SvcIP + ":" + strconv.Itoa(int(edpt.Spec.SvcPort)),
				})
			}
			fmt.Println(table)
		}
	case "replica":
		{
			table, _ := gotable.Create("NAME", "DESIRED", "CURRENT")
			rsList := gjson.Parse(_json).Array()
			for _, p := range rsList {
				rs := &apiobject.ReplicationController{}
				rs.UnMarshalJSON([]byte(p.String()))
				table.AddRow(map[string]string{
					"NAME":    rs.Data.Name,
					"DESIRED": strconv.Itoa(int(rs.Spec.Replicas)),
					"CURRENT": strconv.Itoa(int(rs.Status.Replicas)),
				})
			}
			fmt.Println(table)
		}
	}
	if err != nil {
		//log.Fatal(err)
		/* 解析info，错误判断pod名字是否存在 */
		fmt.Print(_json)
	}

	/* {"error":"key not found: /registry/pods/default/test"} */
	/* {"metadata":{"name":"test-pod"},"spec":{"containers":[{"name":"test-container","resources":{"limits":{},"requests":{}}}]},"status":{"phase":"Pending"}}root@minik8s-2:~/mini-k8s/pkg/kubectl/test# */

}
