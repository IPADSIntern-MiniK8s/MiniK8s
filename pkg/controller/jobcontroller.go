package controller

import (
	log "github.com/sirupsen/logrus"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/utils"
	"time"
)

type jobHandler struct {
}

type jobPodHandler struct {
}

func (r jobHandler) HandleCreate(message []byte) {
	job := &apiobject.Job{}
	job.UnMarshalJSON(message)

	createPod(*job)
	job.Status.Phase = apiobject.Created
	utils.UpdateObject(job,config.JOB,job.Data.Namespace,job.Data.Name)

	log.Info("[job controller] Create job. Name:", job.Data.Name)
}

func (r jobHandler) HandleDelete(message []byte) {
	job := &apiobject.Job{}
	job.UnMarshalJSON(message)

	deletePod(*job)
	utils.DeleteObject(config.JOB, job.Data.Namespace, job.Data.Name)

	log.Info("[job controller] Delete job. Name:", job.Data.Name)
}

func (r jobHandler) HandleUpdate(message []byte) {

}

func (r jobHandler) GetType() config.ObjType {
	return config.JOB
}

/* ========== Start Pod Handler ========== */

func (p jobPodHandler) HandleCreate(message []byte) {

}

func (p jobPodHandler) HandleDelete(message []byte) {
	//delete job-> delete pod
	//not consider pod deleted by user directly
}

func (p jobPodHandler) HandleUpdate(message []byte) {
	pod := &apiobject.Pod{}
	pod.UnMarshalJSON(message)


	if pod.Status.OwnerReference.Kind == config.JOB{
		log.Info("[job controller] handleupdate")
		info := utils.GetObject(config.JOB, pod.Data.Namespace, pod.Status.OwnerReference.Name)
		job:= &apiobject.Job{}
		job.UnMarshalJSON([]byte(info))


		job.Status.Phase = pod.Status.Phase
		switch pod.Status.Phase{
		case apiobject.Failed:{
			job.Spec.BackoffLimit -= 1
			if job.Spec.BackoffLimit > 0{
				go func(job apiobject.Job){
					deletePod(job)
					time.Sleep(time.Second*5)
					createPod(job)
				}(*job)
			}
		}
		case apiobject.Finished:{
			waitToDelete:=func(t int, job apiobject.Job){
				time.Sleep(time.Second * time.Duration(t))
				deletePod(job)
			}
			go waitToDelete(job.Spec.TtlSecondsAfterFinished,*job)

		}
		}
		log.Info("[job controller] update phase:", job.Status.Phase)
		utils.UpdateObject(job, config.JOB, pod.Data.Namespace, pod.Data.Name)
	}

}

func (p jobPodHandler) GetType() config.ObjType {
	return config.POD
}



func createPod(job apiobject.Job){
	pod := &apiobject.Pod{
		Data: job.Data,
		Spec: apiobject.PodSpec{
			NodeSelector:job.Spec.NodeSelector,
			Containers:job.Spec.Containers,
			Volumes:job.Spec.Volumes,
		},
	}
	pod.Status.OwnerReference = apiobject.OwnerReference{
		Kind:       config.JOB,
		Name:       job.Data.Name,
		Controller: false,
	}

	utils.CreateObject(pod,config.POD,job.Data.Namespace)

}
func deletePod(job apiobject.Job){
	utils.DeleteObject(config.POD, job.Data.Namespace, job.Data.Name)

}
