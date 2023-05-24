package controller

import (
	log "github.com/sirupsen/logrus"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/utils"
)

type jobHandler struct {
}

type jobPodHandler struct {
}

func (r jobHandler) HandleCreate(message []byte) {
	job := &apiobject.Job{}
	job.UnMarshalJSON(message)

	pod := &apiobject.Pod{
		Data: job.Data,
		Spec: job.Spec,
	}
	pod.Status.OwnerReference = apiobject.OwnerReference{
		Kind:       config.JOB,
		Name:       job.Data.Name,
		Controller: false,
	}

	utils.CreateObject(pod,config.POD,job.Data.Namespace)

	log.Info("[job controller] Create job. Name:", job.Data.Name)
}

func (r jobHandler) HandleDelete(message []byte) {
	job := &apiobject.Job{}
	job.UnMarshalJSON(message)

	utils.DeleteObject(config.POD, job.Data.Namespace, job.Data.Name)
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

	log.Info("[job controller] handleupdate")

	if pod.Status.OwnerReference.Kind == config.JOB{
		log.Info("[job controller] handleupdate1")
		info := utils.GetObject(config.JOB, pod.Data.Namespace, pod.Status.OwnerReference.Name)
		job:= &apiobject.Job{}
		job.UnMarshalJSON([]byte(info))


		job.Status.Phase = pod.Status.Phase
		log.Info("[job controller] update phase:", job.Status.Phase)
		utils.UpdateObject(job, config.JOB, pod.Data.Namespace, pod.Data.Name)
	}

}

func (p jobPodHandler) GetType() config.ObjType {
	return config.POD
}
