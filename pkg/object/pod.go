package objecttype

type Pod struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	APIVersion string `json:"apiVersion"`

	Spec   PodSpec   `json:"spec,omitempty"`
	Status PodStatus `json:"status,omitempty"`
}

type Meta
type PodSpec struct {
	Containers []Container `json:"containers"`
}

type Container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type PodStatus struct {
	Phase string `json:"phase"`
}
