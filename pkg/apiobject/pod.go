package apiobject

import (
	"encoding/json"
	"fmt"
)

/* an basic example of a pod apiobject:
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: default
  labels:
    app: nginx
spec:
  containers:
  - name: nginx
    image: nginx:latest
    imagePullPolicy: IfNotPresent
    command: ["/bin/sh"]
    args: ["-c", "echo Hello Kubernetes!"]
    env:
      - name: DB_HOST
        value: "localhost"
      - name: DB_PORT
        value: "3306"
	resources:
		limits:
        	cpu: "0.5"
        memory: "250Mi"
      	requests:
        	cpu: "0.25"
			memory: "125Mi"
    ports:
      - containerPort: 80
        name: http
        protocol: TCP
    volumeMounts:
      - name: data
        mountPath: /data
  volumes:
    - name: data
      hostPath:
        path: /data

*/

type Pod struct {
	APIVersion string    `json:"apiVersion,omitempty"`
	Data       MetaData  `json:"metadata"`
	Spec       PodSpec   `json:"spec,omitempty"`
	Status     PodStatus `json:"status,omitempty"`
}

type PodSpec struct {
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Containers   []Container       `json:"containers"`
	Volumes      []Volumes         `json:"volumes,omitempty"`
}

type Container struct {
	Name            string         `json:"name"`
	Image           string         `json:"image,omitempty"`
	ImagePullPolicy string         `json:"imagePullPolicy,omitempty"`
	Command         []string       `json:"command,omitempty"`
	Args            []string       `json:"args,omitempty"`
	Env             []Env          `json:"env,omitempty"`
	Resources       Resources      `json:"resources,omitempty"`
	Ports           []Port         `json:"ports,omitempty"`
	VolumeMounts    []VolumeMounts `json:"volumeMounts,omitempty"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Resources struct {
	Limits   Limit   `json:"limits,omitempty"`
	Requests Request `json:"requests,omitempty"`
}

type Limit struct {
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type Request struct {
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type Port struct {
	ContainerPort int32  `json:"containerPort"`
	Name          string `json:"name,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}

type VolumeMounts struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
}

type Volumes struct {
	Name     string   `json:"name"`
	HostPath HostPath `json:"hostPath"`
}

type HostPath struct {
	Path string `json:"path"`
}

type Volume struct {
	Name string `json:"name"`
}

type PodStatus struct {
	Phase          PhaseLabel     `json:"phase,omitempty""`
	HostIp         string         `json:"hostIP,omitempty"`
	PodIp          string         `json:"podIP,omitempty"`
	OwnerReference OwnerReference `json:"ownerReference,omitempty"`
}

type OwnerReference struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Controller bool   `json:"controller,omitempty"`
}

type PhaseLabel string

const (
	Pending     PhaseLabel = "Pending"
	Running     PhaseLabel = "Running"
	Succeeded   PhaseLabel = "Succeeded"
	Failed      PhaseLabel = "Failed"
	Finished    PhaseLabel = "Finished"
	Terminating PhaseLabel = "Terminating"
	Deleted     PhaseLabel = "Deleted"
	Unknown     PhaseLabel = "Unknown"
)

func (p *Pod) UnMarshalJSON(data []byte) error {
	type Alias Pod
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

func (p *Pod) MarshalJSON() ([]byte, error) {
	type Alias Pod
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	})
}

func (p *Pod) String() string {
	return fmt.Sprintf("Pod: %s", p.Data.Name)
}

func (p *Pod) UnMarshalJsonList(data []byte) ([]Pod, error) {
	var pods []Pod
	err := json.Unmarshal(data, &pods)
	if err != nil {
		return nil, err
	}
	return pods, nil
}
