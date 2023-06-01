package apiobject

import (
	"encoding/json"
)

type Object interface {
	MarshalJSON() ([]byte, error)
	UnMarshalJSON(data []byte) error
}

type MetaData struct {
	Name             string            `json:"name,omitempty"`
	Namespace        string            `json:"namespace,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	ResourcesVersion VersionLabel      `json:"resourcesVersion,omitempty"` // use for update
}

type VersionLabel string

const (
	DELETE VersionLabel = "delete"
	UPDATE VersionLabel = "update"
	CREATE VersionLabel = "create"
)

// MarshalJSONList the object list to json
func MarshalJSONList(list interface{}) ([]byte, error) {
	return json.Marshal(list)
}
