package controller

import (
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"minik8s/pkg/apiobject"
	"minik8s/utils"
)

/*
	主要工作：

1. 监听replication资源的创建。一旦创建，寻找符合selector条件的pod并将其owner设置为当前replicaset。
2. 如果满足条件pod个数小于replicas值，根据template创建新的pod。
3. 监听replication资源的更新。监听selector条件的更改和目标replicas数目和当前replica数目的更改。
4. 监听replication资源的删除。删除对应的pod。
5. 监听pod删除。如果满足controller条件，对应replicas状态数减1，根据template创建新的pod。
6. 监听pod更新：查看label是否更改。并更改对应controller状态。
*/

type rsReplicaHandler struct {
}

type rsPodHandler struct {
}

/* ========== Start Replication Handler ========== */

func (r rsReplicaHandler) HandleCreate(message []byte) {
	rs := &apiobject.ReplicationController{}
	rs.UnMarshalJSON(message)

	// 1. traverse the pod list to find the pod fits the selector and add owner reference to them
	rest := createFromPodList(rs)

	// 2. if the number is less than the required number, create new pods according to the template
	if rest > 0 {
		createFromTemplate(rs.Spec.Template, rest, rs.Data.Name, rs.Data.Namespace)
	}

	rs.Status.Replicas = rs.Spec.Replicas
	utils.UpdateObject(rs, utils.REPLICA, rs.Data.Namespace, rs.Data.Name)

	log.Info("[rs controller] Create replicaset. Name:", rs.Data.Name)
}

func (r rsReplicaHandler) HandleDelete(message []byte) {
	rs := &apiobject.ReplicationController{}
	rs.UnMarshalJSON(message)

	info := utils.GetObject(utils.POD, rs.Data.Namespace, "")
	podList := gjson.Parse(info).Array()
	for _, p := range podList {
		pod := &apiobject.Pod{}
		pod.UnMarshalJSON([]byte(p.String()))
		if isPodBelongToController(pod, rs) {
			utils.DeleteObject(utils.POD, pod.Data.Namespace, pod.Data.Name)
		}
	}
	utils.DeleteObject(utils.REPLICA, rs.Data.Namespace, rs.Data.Name)

	log.Info("[rs controller] Delete replicaset. Name:", rs.Data.Name)
}

func (r rsReplicaHandler) HandleUpdate(message []byte) {
	rs := &apiobject.ReplicationController{}
	rs.UnMarshalJSON(message)

	if rs.Spec.Replicas > rs.Status.Replicas {
		rest := createFromPodList(rs)
		if rest > 0 {
			createFromTemplate(rs.Spec.Template, rest, rs.Data.Name, rs.Data.Namespace)
		}
		rs.Status.Replicas = rs.Spec.Replicas
		utils.UpdateObject(rs, utils.REPLICA, rs.Data.Namespace, rs.Data.Name)
	}

	/* TODO: rs selector change */
}

func (r rsReplicaHandler) GetType() utils.ObjType {
	return utils.REPLICA
}

/* ========== Start Pod Handler ========== */

func (p rsPodHandler) HandleCreate(message []byte) {

}

func (p rsPodHandler) HandleDelete(message []byte) {
	pod := &apiobject.Pod{}
	pod.UnMarshalJSON(message)

	// check if the pod belongs to the replica
	if pod.Status.OwnerReference.Controller && pod.Status.OwnerReference.Kind == "replica" {
		info := utils.GetObject(utils.REPLICA, pod.Data.Namespace, pod.Status.OwnerReference.Name)
		rs := &apiobject.ReplicationController{}
		rs.UnMarshalJSON([]byte(info))
		rs.Status.Replicas = rs.Status.Replicas - 1
		utils.UpdateObject(rs, utils.REPLICA, rs.Data.Namespace, rs.Data.Name)
	}
}

func (p rsPodHandler) HandleUpdate(message []byte) {
	pod := &apiobject.Pod{}
	pod.UnMarshalJSON(message)

	// check if the pod label changes
	info := utils.GetObject(utils.REPLICA, pod.Data.Namespace, pod.Status.OwnerReference.Name)
	rs := &apiobject.ReplicationController{}
	rs.UnMarshalJSON([]byte(info))
	if !isPodBelongToController(pod, rs) {
		// update rs controller
		rs.Status.Replicas = rs.Status.Replicas - 1
		utils.UpdateObject(rs, utils.REPLICA, rs.Data.Namespace, rs.Data.Name)

		// update pod: delete controller info
		pod.Status.OwnerReference.Controller = false
		utils.UpdateObject(pod, utils.REPLICA, pod.Data.Namespace, pod.Data.Name)
	}
}

func (p rsPodHandler) GetType() utils.ObjType {
	return utils.POD
}

/* ========== Util Function ========== */

func createFromPodList(rs *apiobject.ReplicationController) int32 {
	info := utils.GetObject(utils.POD, rs.Data.Namespace, "")
	podList := gjson.Parse(info).Array()
	var num = rs.Status.Replicas
	for _, p := range podList {
		pod := &apiobject.Pod{}
		pod.UnMarshalJSON([]byte(p.String()))
		if pod.Status.OwnerReference.Controller == false && utils.IsLabelEqual(rs.Spec.Selector, pod.Data.Labels) {
			setController(rs.Data.Name, pod)
			utils.UpdateObject(pod, utils.POD, pod.Data.Namespace, pod.Data.Name)
			num++
			if num == rs.Spec.Replicas {
				break
			}
		}
	}
	log.Info("[rs controller] Create from pod list. Create Num:", num-rs.Status.Replicas)
	return rs.Spec.Replicas - num
}

func createFromTemplate(t *apiobject.PodTemplateSpec, num int32, name string, ns string) {
	for i := 0; i < int(num); i++ {
		pod := &apiobject.Pod{
			Data: t.Data,
			Spec: t.Spec,
		}
		pod.Data.Name = utils.GenerateName(name, 10)
		pod.Data.Namespace = ns
		setController(name, pod)
		print("name: ", pod.Data.Name, "\n")
		utils.CreateObject(pod, utils.POD, ns)
	}

	log.Info("[rs controller] Create from template. Create Num:", num)

}

func setController(name string, p *apiobject.Pod) {
	p.Status.OwnerReference = apiobject.OwnerReference{
		Kind:       "replica",
		Name:       name,
		Controller: true,
	}
}

func isPodBelongToController(p *apiobject.Pod, c *apiobject.ReplicationController) bool {
	if p.Status.OwnerReference.Controller == true && p.Status.OwnerReference.Name == c.Data.Name && p.Status.OwnerReference.Kind == "replica" {
		return true
	} else {
		return false
	}
}
