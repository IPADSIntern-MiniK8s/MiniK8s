package utils

import "fmt"

var ApiServerIp = "192.168.1.13:8080"
var httpPrefix = fmt.Sprintf("http//%s/api/v1/", ApiServerIp)

type ObjType string

const (
	POD      ObjType = "pods"
	SERVICE  ObjType = "services"
	ENDPOINT ObjType = "endpoints"
)
