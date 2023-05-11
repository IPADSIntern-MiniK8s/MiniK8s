package ctlutils

import (
	"fmt"
	"github.com/tidwall/gjson"
	"strings"
)

var apiServerIp = "http://127.0.0.1:8080"

var Resources = []string{"pod"}

func ParseUrlFromJson(_json []byte) string {
	// operation: create/apply. eg: POST "/api/v1/namespaces/{namespace}/pods"
	kind := strings.ToLower(gjson.Get(string(_json), "kind").String())
	namespace := gjson.Get(string(_json), "metadata.namespace")

	url := fmt.Sprintf("%s/api/v1/namespaces/%s/%ss", apiServerIp, namespace, kind)
	return url
}

func ParseUrlMany(kind string, ns string) string {
	// operation: get. eg: GET "/api/v1/namespaces/{namespace}/pods"
	// operation: create/apply. eg: POST "/api/v1/namespaces/{namespace}/pods"
	namespace := ns
	url := fmt.Sprintf("%s/api/v1/namespaces/%s/%ss", apiServerIp, namespace, kind)
	return url
}

func ParseUrlOne(kind string, name string, ns string) string {
	// operation: get. eg: "/api/v1/namespaces/{namespace}/pods/{pod_name}"
	namespace := ns
	url := fmt.Sprintf("%s/api/v1/namespaces/%s/%ss/%s", apiServerIp, namespace, kind, name)
	return url
}
