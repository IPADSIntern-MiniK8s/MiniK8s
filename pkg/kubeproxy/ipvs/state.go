package ipvs

import "github.com/mqliang/libipvs"

type ServiceNode struct {
	service   *libipvs.Service
	endpoints []*libipvs.Destination
}

var Services map[string]ServiceNode
var Endpoints map[string]*libipvs.Destination
