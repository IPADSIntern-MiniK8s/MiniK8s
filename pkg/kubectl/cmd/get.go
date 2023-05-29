package cmd

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/apiobject"
	ctlutils "minik8s/pkg/kubectl/utils"
	"minik8s/utils"
	"strconv"
	"strings"

	"github.com/liushuochen/gotable"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/wxnacy/wgo/arrays"
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
		/* get all resources in certain type under specified namespace */
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
			table, _ := gotable.Create("NAME", "POD-IP", "STATUS", "NODE-IP")
			podList := gjson.Parse(_json).Array()
			for _, p := range podList {
				name := gjson.Get(p.String(), "metadata.name").String()
				status := gjson.Get(p.String(), "status.phase").String()
				IP := gjson.Get(p.String(), "status.podIP").String()
				nodeIP := gjson.Get(p.String(), "status.hostIP").String()
				table.AddRow(map[string]string{
					"NAME":    name,
					"POD-IP":  IP,
					"STATUS":  status,
					"NODE-IP": nodeIP,
				})
			}
			fmt.Println(table)
		}
	case "job":
		{
			table, _ := gotable.Create("NAME", "POD-NAME", "STATUS")
			job := gjson.Parse(_json).Array()
			for _, p := range job {
				name := gjson.Get(p.String(), "metadata.name").String()
				status := gjson.Get(p.String(), "status.phase").String()
				table.AddRow(map[string]string{
					"NAME":     name,
					"POD-NAME": name,
					"STATUS":   status,
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
					"PORT(S)":    portString,
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
			table, _ := gotable.Create("NAME", "DESIRED", "READY")
			rsList := gjson.Parse(_json).Array()
			for _, p := range rsList {
				rs := &apiobject.ReplicationController{}
				rs.UnMarshalJSON([]byte(p.String()))
				table.AddRow(map[string]string{
					"NAME":    rs.Data.Name,
					"DESIRED": strconv.Itoa(int(rs.Spec.Replicas)),
					"READY":   strconv.Itoa(int(rs.Status.ReadyReplicas)),
				})
			}
			fmt.Println(table)
		}
	case "hpa":
		{
			table, _ := gotable.Create("NAME", "REFERENCE", "TARGETS", "MINPODS", "MAXPODS")
			hpaList := gjson.Parse(_json).Array()
			for _, p := range hpaList {
				hpa := &apiobject.HorizontalPodAutoscaler{}
				hpa.UnMarshalJSON([]byte(p.String()))
				target := ""
				for i, m := range hpa.Spec.Metrics {
					if i < len(hpa.Status.CurrentMetrics) {
						target += strconv.Itoa(hpa.GetStatusValue(&hpa.Status.CurrentMetrics[i])) + "/" + strconv.Itoa(hpa.GetTargetValue(&m)) + ","
					}
				}
				table.AddRow(map[string]string{
					"NAME":      hpa.Data.Name,
					"REFERENCE": string(hpa.Spec.ScaleTargetRef.Kind) + "/" + hpa.Spec.ScaleTargetRef.Name,
					"TARGETS":   target,
					"MINPODS":   strconv.Itoa(int(hpa.Spec.MinReplicas)),
					"MAXPODS":   strconv.Itoa(int(hpa.Spec.MaxReplicas)),
				})
			}
			fmt.Println(table)
		}
	case "function":
		{
			table, _ := gotable.Create("NAME", "PATH", "STATUS")
			funcList := gjson.Parse(_json).Array()
			for _, f := range funcList {
				function := &apiobject.Function{}
				function.UnMarshalJSON([]byte(f.String()))
				table.AddRow(map[string]string{
					"NAME":   function.Name,
					"PATH":   function.Path,
					"STATUS": string(function.Status),
				})
			}
			fmt.Println(table)
		}
	case "workflow":
		{
			table, _ := gotable.Create("NAME", "STATUS")
			wfList := gjson.Parse(_json).Array()
			for _, f := range wfList {
				wf := &apiobject.WorkFlow{}
				wf.UnMarshalJSON([]byte(f.String()))
				table.AddRow(map[string]string{
					"NAME":   wf.Name,
					"STATUS": string(wf.Status),
				})
			}
		}
	case "node":
		{
			table, _ := gotable.Create("NAME", "IP", "STATUS")
			nodeList := gjson.Parse(_json).Array()
			for _, f := range nodeList {
				node := &apiobject.Node{}
				node.UnMarshalJSON([]byte(f.String()))
				table.AddRow(map[string]string{
					"NAME":   node.Data.Name,
					"IP":     node.Status.Addresses[0].Address,
					"STATUS": string(node.Status.Conditions[0].Status),
				})
			}
		}
	case "dnsrecord":
		{
			table, _ := gotable.Create("NAME", "HOST", "PATHS")
			dnsList := gjson.Parse(_json).Array()
			for _, f := range dnsList {
				dns := &apiobject.DNSRecord{}
				dns.UnMarshalJSON([]byte(f.String()))
				jsonBytes, _ := json.Marshal(dns.Paths)
				table.AddRow(map[string]string{
					"NAME":  dns.Name,
					"HOST":  dns.Host,
					"PATHS": string(jsonBytes),
				})
			}
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
