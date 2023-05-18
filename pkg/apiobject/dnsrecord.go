package apiobject

// example:

import "encoding/json"

type DNSRecord struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Paths      []Path `json:"paths"`
}

type Path struct {
	Address string `json:"address,omitempty"`
	Service string `json:"service"`
	Port    int    `json:"port"`
}

type DNSEntry struct {
	Host string `json:"host"`
}

func (r *DNSRecord) MarshalJSON() ([]byte, error) {
	type Alias DNSRecord
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

func (r *DNSRecord) UnMarshalJSON(data []byte) error {
	type Alias DNSRecord
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
