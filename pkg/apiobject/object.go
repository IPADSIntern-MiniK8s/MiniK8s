package apiobject

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	OwnerReference   OwnerReference    `json:"OwnerReference,omitempty"`
}

type OwnerReference struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Controller *bool  `json:"controller,omitempty"`
}

type VersionLabel string

const (
	DELETE VersionLabel = "delete"
	UPDATE VersionLabel = "update"
	CREATE VersionLabel = "create"
)

func UnMarshalJSONList(jsonData []byte, dest interface{}) error {
	err := json.Unmarshal(jsonData, dest)
	if err != nil {
		return err
	}

	// check whether the result is a slice
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Slice {
		return fmt.Errorf("result is not a slice")
	}

	// traverse the slice and check whether all elements are of the same type
	elementType := value.Type().Elem()
	for i := 0; i < value.Len(); i++ {
		element := value.Index(i)
		if !element.Type().AssignableTo(elementType) {
			return fmt.Errorf("element %d is not of type %s", i, elementType.Name())
		}
	}

	return nil
}

// MarshalJSONList the object list to json
func MarshalJSONList(list interface{}) ([]byte, error) {
	return json.Marshal(list)
}
