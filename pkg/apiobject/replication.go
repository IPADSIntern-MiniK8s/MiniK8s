package apiobject

import "encoding/json"

/*
	an basic example of a repicaset apiobject:

kind: Deployment
apiVersion: apps/v1
metadata:

	name: deploy-practice

spec:

	replicas: 3
	selector:
	    app: deploy-practice
	template:
	  metadata:
	    labels:
	      app: deploy-practice
	  spec:
	    containers:
	    - name: fileserver
	      image: hejingkai/fileserver:latest
	      ports:
	      - name: p1 # 端口名称
	        containerPort: 8080  # 容器端口
	      volumeMounts:
	      - name: download
	        mountPath: /usr/share/files
	    - name: downloader
	      image: hejingkai/downloader:latest
	      ports:
	      - name: p2 # 端口名称
	        containerPort: 3000  # 容器端口
	      volumeMounts:
	      - name: download
	        mountPath: /data
	    volumes: # 定义数据卷
	    - name: download # 数据卷名称
	      emptyDir: {} # 数据卷类型
*/
type ReplicationController struct {
	Kind       string                      `json:"kind,omitempty"`
	APIVersion string                      `json:"apiVersion,omitempty"`
	Data       MetaData                    `json:"metadata"`
	Spec       ReplicationControllerSpec   `json:"spec,omitempty"`
	Status     ReplicationControllerStatus `json:"status,omitempty"`
}

type ReplicationControllerSpec struct {
	// Replicas is the number of desired replicas.
	Replicas int32 `json:"replicas"`

	// Selector is a label query over pods that should match the Replicas count.
	Selector map[string]string `json:"selector"`

	Template *PodTemplateSpec `json:"template"`
}

type PodTemplateSpec struct {
	Data MetaData `json:"metadata"`
	Spec PodSpec  `json:"spec"`
}

type ReplicationControllerStatus struct {
	// Replicas is the number of actual replicas.
	Replicas int32 `json:"replicas"`
	// Used for replica controlled by HPA
	Scale          int32          `json:"scale"`
	OwnerReference OwnerReference `json:"ownerReference,omitempty"`
	// the truly ready replicas.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
}

func (r *ReplicationController) UnMarshalJSON(data []byte) error {
	type Alias ReplicationController
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

func (r *ReplicationController) MarshalJSON() ([]byte, error) {
	type Alias ReplicationController
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

func (r *ReplicationController) Union(other *ReplicationController) {
	if r.Status.Replicas == 0 {
		r.Status.Replicas = other.Status.Replicas
	}
	empty := OwnerReference{}
	if empty == r.Status.OwnerReference {
		r.Status.OwnerReference = other.Status.OwnerReference
	}
}

func (r *ReplicationController) UnMarshalJSONList(data []byte) ([]ReplicationController, error) {
	var list []ReplicationController
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	return list, nil
}
