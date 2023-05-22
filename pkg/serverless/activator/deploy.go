package activator

import (
	"bytes"
	"errors"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/controller"
	"minik8s/pkg/serverless/autoscaler"
	"minik8s/pkg/serverless/function"
	"minik8s/utils"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

const retryTimes = 3
const serverIp = "master"



func GenerateReplicaSet(name string, namespace string, image string, replicas int32) *apiobject.ReplicationController {
	return &apiobject.ReplicationController {
		Kind:       "ReplicaSet",
		APIVersion: "apps/v1",
		Data: apiobject.MetaData{
			Name:      name,
			Namespace: namespace,
		},
		Spec: apiobject.ReplicationControllerSpec {
			Replicas: replicas,
			Selector: map[string]string {
				"app": name,
			},
			Template: &apiobject.PodTemplateSpec {
				Data: apiobject.MetaData{
					Name:      name,
					Namespace: namespace,
					Labels: map[string]string {
						"app": name,
					},
				}, 
				Spec: apiobject.PodSpec {
					Containers: []apiobject.Container {
						{
							Name:  name,
							Image: image,
							Ports: []apiobject.Port {
								{
									ContainerPort: 8081,
									Name: "p1",
									Protocol: "TCP",
								},
							},
						},
					},
				},
			},
		},
		Status: apiobject.ReplicationControllerStatus {
			Replicas: 0,
			Scale: 0,
			OwnerReference: apiobject.OwnerReference {
				Kind: config.FUNCTION,
				Name: name,
				Controller: true,
			},
		},
	}
}

func getPodIpList(pods []*apiobject.Pod) []string {
	result := make([]string, 0)
	for _, pod := range pods {
		result = append(result, pod.Status.PodIp)
	}
	return result
}

// CheckPrepare check if the function is deployed
// if not, deploy it
// if yes, return the pod ip
// wait until the function is deployed or timeout
// return the pod ips
func CheckPrepare(name string) ([]string, error) {
	// 1. find the according replicaSet
	replicaUrl := "http://" + config.ApiServerIp + "/api/v1/namespaces/serverless/replicas/" + name
	response, err := utils.SendRequest("GET", nil, replicaUrl)
	if err != nil {
		log.Error("[CheckPrepare] error getting replicas: ", err)
		return nil, err
	}
	replicaSet := &apiobject.ReplicationController{}
	err = replicaSet.UnMarshalJSON([]byte(response))
	if err != nil {
		log.Error("[CheckPrepare] error unmarshalling replicas: ", err)
		return nil, err
	}

	// 2. check the deployment status
	// retry for 3 times
	retry := retryTimes
	for retry > 0 {
		timer := time.NewTimer(10 * time.Minute)
		deployed := false

		for {
			select {
			case <- timer.C:
				break
			default:
				if !deployed {
					// the first time, check if the function is deployed
					pods := controller.GetPodListFromRS(replicaSet)
					if len(pods) == 0 {
						// deployed it now
						replicaSet.Status.Scale = 1
						utils.UpdateObject(replicaSet, config.REPLICA, replicaSet.Data.Namespace, replicaSet.Data.Name)
					} else {
						// generate the pod ip list
						podIps := getPodIpList(pods)
						autoscaler.RecordMutex.Lock()
						record:= autoscaler.GetRecord(name)
						if record == nil {
							autoscaler.RecordMap[name] = &autoscaler.Record {
								Name : name,
								Replicas: replicaSet.Status.Replicas,
								PodIps: make(map[string]int32),
								CallCount: 1,
							}
							autoscaler.RecordMutex.Unlock()
							return podIps, nil
						} else {
							record.CallCount++
							record.Replicas = int32(len(podIps))
							autoscaler.RecordMap[name] = record
							// if the call count is larger than the threshold, scale up
							if record.CallCount > record.Replicas {
								replicaSet.Status.Scale = record.CallCount + 1
								utils.UpdateObject(replicaSet, config.REPLICA, replicaSet.Data.Namespace, replicaSet.Data.Name)
							} else {
								autoscaler.RecordMutex.Unlock()
								return podIps, nil
							}
						}
						autoscaler.RecordMutex.Unlock()
					}
				} else {
					// check whether the function is deployed and the replica number is correct
					pods := controller.GetPodListFromRS(replicaSet)
					autoscaler.RecordMutex.RLock()
					record := autoscaler.GetRecord(name)
					autoscaler.RecordMutex.RUnlock()
					if record == nil {
						return nil, errors.New("record not found")
					} 
					
					if int32(len(pods)) == record.Replicas {
						podsIp := getPodIpList(pods)
						return podsIp, nil
					}
				}

				time.Sleep(3 * time.Second)
			}
			retry--
		}
	}
	
	// get the current pod ip list and return
	pods := controller.GetPodListFromRS(replicaSet)
	podsIp := getPodIpList(pods)
	return podsIp, nil
}


// InitFunc init the function, initialize the replicaSet and generate the image
func InitFunc(name string, path string) error {
	err := function.CreateImage(path, name)
	if err != nil {
		log.Error("[InitFunc] create image error: ", err)
		return err
	}
	ImageName := serverIp + ":5000/" + name + ":latest"
	replicaSet := GenerateReplicaSet(name, "serverless", ImageName, 0)

	// create the record 
	autoscaler.RecordMutex.Lock()
	autoscaler.RecordMap[name] = &autoscaler.Record {
		Name : name,
		Replicas: 0,
		PodIps: make(map[string]int32),
		CallCount: 0,
	}
	autoscaler.RecordMutex.Unlock()

	utils.CreateObject(replicaSet, config.REPLICA, replicaSet.Data.Namespace)
	return nil
}


// LoadBalance choose a pod ip to trigger the function
func LoadBalance(name string, podIps []string) (string, error) {
	if len(podIps) == 0 {
		log.Error("[LoadBalance] pod ip list is empty")
		return "", errors.New("pod ip list is empty")
	}

	autoscaler.RecordMutex.RLock()
	record := autoscaler.GetRecord(name)
	autoscaler.RecordMutex.RUnlock()

	if record == nil { 
		log.Error("[LoadBalance] record not found")
		return "", errors.New("record not found")
	}

	// update the record
	for _, podIp := range podIps {
		if _, ok := record.PodIps[podIp]; !ok {
			record.PodIps[podIp] = 0
		}
	}

	// choose the pod ip with the least call count
	sort.Slice(podIps, func(i, j int) bool {
		return record.PodIps[podIps[i]] < record.PodIps[podIps[j]]
	})

	chosenPodIp := podIps[0]
	record.PodIps[chosenPodIp]++

	autoscaler.RecordMutex.Lock()
	autoscaler.RecordMap[name] = record
	autoscaler.RecordMutex.Unlock()
	
	return chosenPodIp, nil
}


// TriggerFunc trigger the function with some parameters
// if the function is not deployed, deploy it first
func TriggerFunc(name string, params []byte) ([]byte, error) {
	// 1. check if the function is deployed
	podIps, err := CheckPrepare(name)
	if err != nil {
		log.Error("[TriggerFunc] check prepare error: ", err)
		return nil, err
	}

	// 2. load balance
	podIp, err := LoadBalance(name, podIps)
	if err != nil {
		log.Error("[TriggerFunc] load balance error: ", err)
		return nil, err
	}

	// 3. trigger the function
	url := "http://" + podIp + ":8081/"
	ret, err := utils.SendRequest("POST", params, url)
	result := bytes.NewBufferString(ret).Bytes()
	if err != nil {
		log.Error("[TriggerFunc] send request error: ", err)
		return nil, err
	}
	
	return result, err
}


// DeleteFunc delete the function
func DeleteFunc(name string) error {
	// TODO: how to delete replicaset?
	// 1. delete the replicaset
	replicaUrl := "http://" + config.ApiServerIp + "/api/v1/namespaces/serverless/replicas/" + name
	_, err := utils.SendRequest("DELETE", nil, replicaUrl)
	if err != nil {
		log.Error("[DeleteFunc] delete replicas error: ", err)
		return err
	}

	// 2. delete the record from the record map
	autoscaler.RecordMutex.Lock()
	delete(autoscaler.RecordMap, name)
	autoscaler.RecordMutex.Unlock()

	// 3. delete the image
	err = function.DeleteImage(name)
	if err != nil {
		log.Error("[DeleteFunc] delete image error: ", err)
		return err
	}
	return nil
}