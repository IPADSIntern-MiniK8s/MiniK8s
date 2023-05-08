package apiobject

import (
	"encoding/json"
)

type Endpoint struct {
	Data MetaData     `json:"metadata"`
	Spec EndpointSpec `json:"spec"`
}

type EndpointSpec struct {
	SvcIP    string `json:"svcIP"`
	SvcPort  int32  `json:"svcPort"`
	DestIP   string `json:"dstIP"`
	DestPort int32  `json:"dstPort"`
}

func (e *Endpoint) MarshalJSON() ([]byte, error) {
	type Alias Endpoint
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	})
}

func (e *Endpoint) UnMarshalJSON(data []byte) error {
	type Alias Endpoint
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}
