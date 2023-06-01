package controller

import (
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/utils"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
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

	var expectReplica int32
	// check if the rs is controlled
	if rs.Status.OwnerReference.Controller == true {
		expectReplica = rs.Status.Scale
	} else {
		expectReplica = rs.Spec.Replicas
	}

	// 1. traverse the pod list to find the pod fits the selector and add owner reference to them
	rest := createFromPodList(rs, expectReplica)

	// 2. if the number is less than the required number, create new pods according to the template
	if rest > 0 {
		createFromTemplate(rs.Spec.Template, rest, rs.Data.Name, rs.Data.Namespace)
	}

	rs.Status.Replicas = expectReplica
	utils.UpdateObject(rs, config.REPLICA, rs.Data.Namespace, rs.Data.Name)

	log.Info("[rs controller] Create replicaset. Name:", rs.Data.Name)
}

func (r rsReplicaHandler) HandleDelete(message []byte) {
	rs := &apiobject.ReplicationController{}
	rs.UnMarshalJSON(message)

	deleteFromPodList(rs, rs.Status.Replicas)

	log.Info("[rs controller] Delete replicaset. Name:", rs.Data.Name)
}

func (r rsReplicaHandler) HandleUpdate(message []byte) {
	rs := &apiobject.ReplicationController{}
	rs.UnMarshalJSON(message)

	var expectReplica int32
	// check if the rs is controlled by HPA
	if rs.Status.OwnerReference.Controller == true {
		expectReplica = rs.Status.Scale
	} else {
		expectReplica = rs.Spec.Replicas
	}
	if expectReplica > rs.Status.Replicas {
		rest := createFromPodList(rs, expectReplica)
		if rest > 0 {
			createFromTemplate(rs.Spec.Template, rest, rs.Data.Name, rs.Data.Namespace)
		}
		rs.Status.Replicas = expectReplica
		utils.UpdateObject(rs, config.REPLICA, rs.Data.Namespace, rs.Data.Name)
	} else if expectReplica < rs.Status.Replicas {
		// choose  pod to delete
		deleteFromPodList(rs, rs.Status.Replicas-expectReplica)
		rs.Status.Replicas = expectReplica
		utils.UpdateObject(rs, config.REPLICA, rs.Data.Namespace, rs.Data.Name)
	}

	/* TODO: rs selector change */
}

func (r rsReplicaHandler) GetType() config.ObjType {
	return config.REPLICA
}

/* ========== Start Pod Handler ========== */

func (p rsPodHandler) HandleCreate(message []byte) {

}

func (p rsPodHandler) HandleDelete(message []byte) {
	pod := &apiobject.Pod{}
	pod.UnMarshalJSON(message)

	// check if the pod belongs to the replica
	if pod.Status.OwnerReference.Controller && pod.Status.OwnerReference.Kind == config.REPLICA {
		info := utils.GetObject(config.REPLICA, pod.Data.Namespace, pod.Status.OwnerReference.Name)
		rs := &apiobject.ReplicationController{}
		rs.UnMarshalJSON([]byte(info))
		rs.Status.Replicas = rs.Status.Replicas - 1
		utils.UpdateObject(rs, config.REPLICA, rs.Data.Namespace, rs.Data.Name)
	}
}

func (p rsPodHandler) HandleUpdate(message []byte) {
	pod := &apiobject.Pod{}
	pod.UnMarshalJSON(message)

	if pod.Status.OwnerReference.Controller {
		info := utils.GetObject(config.REPLICA, pod.Data.Namespace, pod.Status.OwnerReference.Name)
		rs := &apiobject.ReplicationController{}
		rs.UnMarshalJSON([]byte(info))

		// check if the pod label changes
		if !utils.IsLabelEqual(rs.Spec.Selector, pod.Data.Labels) {
			// update rs controller
			rs.Status.Replicas = rs.Status.Replicas - 1
			utils.UpdateObject(rs, config.REPLICA, rs.Data.Namespace, rs.Data.Name)

			// update pod: delete controller info
			pod.Status.OwnerReference.Controller = false
			utils.UpdateObject(pod, config.POD, pod.Data.Namespace, pod.Data.Name)
		}

		// check if the pod status if FAILED or FINISHED
		if pod.Status.Phase == apiobject.Failed || pod.Status.Phase == apiobject.Finished {
			utils.DeleteObject(config.POD, pod.Data.Namespace, pod.Data.Name)
		}
	}

}

func (p rsPodHandler) GetType() config.ObjType {
	return config.POD
}

/* ========== Util Function ========== */

func createFromPodList(rs *apiobject.ReplicationController, expect int32) int32 {
	info := utils.GetObject(config.POD, rs.Data.Namespace, "")
	podList := gjson.Parse(info).Array()
	var num = rs.Status.Replicas
	for _, p := range podList {
		pod := &apiobject.Pod{}
		pod.UnMarshalJSON([]byte(p.String()))
		if pod.Status.OwnerReference.Controller == false && utils.IsLabelEqual(rs.Spec.Selector, pod.Data.Labels) {
			setController(rs.Data.Name, pod)
			utils.UpdateObject(pod, config.POD, pod.Data.Namespace, pod.Data.Name)
			num++
			if num == expect {
				break
			}
		}
	}
	log.Info("[rs controller] Create from pod list. Create Num:", num-rs.Status.Replicas)
	return expect - num
}

func deleteFromPodList(rs *apiobject.ReplicationController, num int32) {
	info := utils.GetObject(config.POD, rs.Data.Namespace, "")
	podList := gjson.Parse(info).Array()
	dNum := num
	for _, p := range podList {
		pod := &apiobject.Pod{}
		pod.UnMarshalJSON([]byte(p.String()))
		if isPodBelongToController(pod, rs) {
			pod.Status.OwnerReference.Controller = false
			utils.UpdateObject(pod, config.POD, pod.Data.Namespace, pod.Data.Name)
			utils.DeleteObject(config.POD, pod.Data.Namespace, pod.Data.Name)
			num--
			if num == 0 {
				break
			}
		}
	}
	log.Info("[rs controller] Delete from pod list. Delete Num:", dNum)
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
		print("[rs controller] name: ", pod.Data.Name, "\n")
		utils.CreateObject(pod, config.POD, ns)
	}

	log.Info("[rs controller] Create from template. Create Num:", num)

}

func setController(name string, p *apiobject.Pod) {
	p.Status.OwnerReference = apiobject.OwnerReference{
		Kind:       config.REPLICA,
		Name:       name,
		Controller: true,
	}
}

func isPodBelongToController(p *apiobject.Pod, c *apiobject.ReplicationController) bool {
	if p.Status.OwnerReference.Controller == true && p.Status.OwnerReference.Name == c.Data.Name && p.Status.OwnerReference.Kind == config.REPLICA {
		return true
	} else {
		return false
	}
}

func GetPodListFromRS(rs *apiobject.ReplicationController) []*apiobject.Pod {
	var podList []*apiobject.Pod
	info := utils.GetObject(config.POD, rs.Data.Namespace, "")
	pList := gjson.Parse(info).Array()
	for _, p := range pList {
		pod := &apiobject.Pod{}
		pod.UnMarshalJSON([]byte(p.String()))

		if isPodBelongToController(pod, rs) {
			podList = append(podList, pod)
		}
	}
	return podList
}
