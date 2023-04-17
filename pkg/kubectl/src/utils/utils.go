package ctlutils

import (
	"fmt"
	"github.com/tidwall/gjson"
	"strings"
)

var apiServerIp = "http://127.0.0.1:8080"

func ParseUrlFromJson(_json []byte) string {
	// operation: create/apply. eg: "/api/v1/namespaces/{namespace}/pods"
	// extension: add type to args
	kind := strings.ToLower(gjson.Get(string(_json), "kind").String())
	namespace := gjson.Get(string(_json), "metadata.namespace")

	url := fmt.Sprintf("%s/api/v1/namespaces/%s/%ss", apiServerIp, namespace, kind)
	return url
}
