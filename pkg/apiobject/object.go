package apiobject

type Object interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

type MetaData struct {
	Name             string `json:"name"`
	Namespace        string `json:"namespace,omitempty"`
	Labels           Label  `json:"labels,omitempty"`
	ResourcesVersion string `json:"resourcesVersion,omitempty"` // use for update
}
