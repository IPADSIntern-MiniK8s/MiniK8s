package apiobject

import (
	"encoding/json"
	"fmt"
)

/* an basic example of a job apiobject:
	apiVersion: v1
kind: Pod
metadata:
  name: gpu-job
  namespace: gpu
spec:
  containers:
    - name: gpu-server
      image: gpu-server
      command:
        - "./job.py"
      env:
        - name: source-path
          value: /gpu
        - name: job-name
          value: gpu-matrix
        - name: partition
          value: dgx2
        - name: "N"
          value: "1"
        - name: ntasks-per-node
          value: "1"
        - name: cpus-per-task
          value: "6"
        - name: gres
          value: gpu:1
      volumeMounts:
        - name: share-data
          mountPath: /gpu
  volumes:
    - name: share-data
      hostPath:
        path: /minik8s-sharedata/gpu/matrix


*/

type Job struct {
	APIVersion string    `json:"apiVersion,omitempty"`
	Data       MetaData  `json:"metadata"`
	Spec       PodSpec   `json:"spec,omitempty"`
	Status     PodStatus `json:"status,omitempty"`
}


func (j *Job) UnMarshalJSON(data []byte) error {
	type Alias Job
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(j),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

func (j *Job) MarshalJSON() ([]byte, error) {
	type Alias Job
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(j),
	})
}

func (j *Job) String() string {
	return fmt.Sprintf("Job: %s", j.Data.Name)
}

func (j *Job) UnMarshalJsonList(data []byte) ([]Job, error) {
	var jobs []Job
	err := json.Unmarshal(data, &jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (j *Job) Union(other *Job) {
	if j.Status.Phase == "" {
		j.Status.Phase = other.Status.Phase
	}
}
