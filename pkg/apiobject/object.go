package apiobject

type Object interface {
}

type MetaData struct {
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	ResourcesVersion string            `json:"resourcesVersion,omitempty"` // use for update
}
