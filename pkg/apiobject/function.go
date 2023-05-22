package apiobject

import "encoding/json"

type Function struct {
	Kind 	 	string `json:"kind,omitempty"`
	APIVersion 	string `json:"apiVersion,omitempty"`
	Status 	 	VersionLabel `json:"status,omitempty"`
	Name       	string `json:"name"`
	Path 	 	string `json:"path"`
}



func (r *Function) MarshalJSON() ([]byte, error) {
	type Alias Function
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

func (r *Function) UnMarshalJSON(data []byte) error {
	type Alias Function
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}