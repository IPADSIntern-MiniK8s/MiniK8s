package ctlutils

import (
	"fmt"
	"minik8s/config"
	"strings"

	"github.com/tidwall/gjson"
)

//var apiServerIp = "http://192.168.1.13:8080"

var Resources = []string{"pod", "service", "endpoint", "replica", "job","hpa", "function"}

func ParseUrlFromJson(_json []byte) string {
	// operation: create/apply. eg: POST "/api/v1/namespaces/{namespace}/pods"
	kind := strings.ToLower(gjson.Get(string(_json), "kind").String())
	namespace := gjson.Get(string(_json), "metadata.namespace")

	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%ss", config.ApiServerIp, namespace, kind)
	return url
}

func ParseUrlMany(kind string, ns string) string {
	// operation: get. eg: GET "/api/v1/namespaces/{namespace}/pods"
	// operation: create/apply. eg: POST "/api/v1/namespaces/{namespace}/pods"
	var namespace string
	if ns == "" {
		namespace = "default"
	} else {
		namespace = ns
	}
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%ss", config.ApiServerIp, namespace, kind)
	return url
}

func ParseUrlOne(kind string, name string, ns string) string {
	// operation: get. eg: "/api/v1/namespaces/{namespace}/pods/{pod_name}"
	var namespace string
	if ns == "" {
		namespace = "default"
	} else {
		namespace = ns
	}
	url := fmt.Sprintf("http://%s/api/v1/namespaces/%s/%ss/%s", config.ApiServerIp, namespace, kind, name)
	return url
}
