package activator

import (
	"bytes"
	"encoding/json"
	"errors"
	"minik8s/config"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/controller"
	"minik8s/pkg/serverless/autoscaler"
	"minik8s/pkg/serverless/function"
	"minik8s/utils"
	"net"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

const retryTimes = 10
const serverIp = "master"
const threshold = 6

func GenerateReplicaSet(name string, namespace string, image string, replicas int32) *apiobject.ReplicationController {
	return &apiobject.ReplicationController{
		Kind:       "ReplicaSet",
		APIVersion: "apps/v1",
		Data: apiobject.MetaData{
			Name:      name,
			Namespace: namespace,
		},
		Spec: apiobject.ReplicationControllerSpec{
			Replicas: replicas,
			Selector: map[string]string{
				"app": name,
			},
			Template: &apiobject.PodTemplateSpec{
				Data: apiobject.MetaData{
					Name:      name,
					Namespace: namespace,
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: apiobject.PodSpec{
					Containers: []apiobject.Container{
						{
							Name:  name,
							Image: image,
							Ports: []apiobject.Port{
								{
									ContainerPort: 8081,
									Name:          "p1",
									Protocol:      "TCP",
								},
							},
							Command: []string{
								"python3",
								"server.py",
							},
						},
					},
				},
			},
		},
		Status: apiobject.ReplicationControllerStatus{
			Replicas: 0,
			Scale:    0,
			OwnerReference: apiobject.OwnerReference{
				Kind:       config.FUNCTION,
				Name:       name,
				Controller: true,
			},
		},
	}
}

func getPodIpList(pods []*apiobject.Pod) []string {
	result := make([]string, 0)
	if pods == nil {
		return result
	}
	for _, pod := range pods {
		if pod.Status.Phase != apiobject.Pending && pod.Status.PodIp != "" {
			result = append(result, pod.Status.PodIp)
		}
	}
	return result
}


// CheckConnection check if the connection is ok
func CheckConnection(ip string) error {
	timer := time.NewTimer(30 * time.Second)
	for {
		select {
			case <-timer.C: 
				return errors.New("timeout")
			default: {
				// try to connect to the ip
				address := ip + ":8081"
				conn, err := net.DialTimeout("tcp", address, time.Second)
				if err != nil {
					continue
				}

				defer conn.Close()

				log.Info("[CheckConnection] Connection is ok")
				return nil
			}
		}
		
	}
}

// CheckPrepare check if the function is deployed
// if not, deploy it
// if yes, return the pod ip
// wait until the function is deployed or timeout
// return the pod ips
func CheckPrepare(name string) ([]string, error) {
	log.Info("[CheckPrepare] check prepare for function: ", name)
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
	theFirstTime := true
	for retry > 0 {
		timer := time.NewTimer(60 * time.Second)
		deployed := false
		retry -= 1
		for {
			select {
			case <-timer.C:
				log.Info("[CheckPrepare] timeout")
				break
			default:
				if !deployed {
					// the first time, check if the function is deployed
					log.Info("the first time, check if the function is deployed")
					pods := controller.GetPodListFromRS(replicaSet)
					// generate the pod ip list
					podIps := getPodIpList(pods)

					autoscaler.RecordMutex.Lock()
					record := autoscaler.GetRecord(name)
					if record == nil {
						autoscaler.RecordMap[name] = &autoscaler.Record{
							Name:      name,
							Replicas:  int32(len(podIps)),
							PodIps:    make(map[string]int32),
							CallCount: 1,
						}
						record = autoscaler.RecordMap[name]
					} else {
						if theFirstTime {
							record.CallCount++
							theFirstTime = false
						}
						record.Replicas = int32(len(podIps))
						autoscaler.RecordMap[name] = record
					}
					// if the call count is larger than the threshold, scale up
					log.Info("[CheckPrepare] record found, the call count: ", record.CallCount, "the replica number: ", record.Replicas)
					if record.CallCount > replicaSet.Status.Scale && record.CallCount < threshold {
						replicaSet.Status.Scale = record.CallCount
						log.Info("[CheckPrepare] scale up the function: ", name, "the replica number: ", replicaSet.Status.Scale)
						utils.UpdateObject(replicaSet, config.REPLICA, replicaSet.Data.Namespace, replicaSet.Data.Name)
					} else {
						autoscaler.RecordMutex.Unlock()
						if len(podIps) > 0 {
							return podIps, nil
						}
					}
					autoscaler.RecordMutex.Unlock()
					deployed = true
				} else {
					// check whether the function is deployed and the replica number is correct
					pods := controller.GetPodListFromRS(replicaSet)
					autoscaler.RecordMutex.RLock()
					record := autoscaler.GetRecord(name)
					autoscaler.RecordMutex.RUnlock()
					if record == nil {
						log.Error("[CheckPrepare] record not found")
						return nil, errors.New("record not found")
					}

					podsIp := getPodIpList(pods)
					log.Info("[CheckPrepare] the pod ip list in second time or later: ", podsIp)
					log.Info("the replica number: ", int32(len(podsIp)), " the call count: ", record.CallCount)
					if (int32(len(podsIp)) >= record.CallCount && len(podsIp) > 0) {
						// update the replica count
						record.Replicas = int32(len(pods))
						autoscaler.RecordMutex.Lock()
						autoscaler.RecordMap[name] = record
						autoscaler.RecordMutex.Unlock()
						log.Info("[CheckPrepare] the replica number is correct: ", int32(len(podsIp)), record.CallCount)
						return podsIp, nil
					}
				}

				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	// get the current pod ip list and return
	log.Info("[CheckPrepare] get the current pod ip list and return")
	pods := controller.GetPodListFromRS(replicaSet)
	podsIp := getPodIpList(pods)
	return podsIp, nil
}

// InitFunc init the function, initialize the replicaSet and generate the image
func InitFunc(name string, path string) error {
	// check whether the function name is duplicated
	info := utils.GetObject(config.REPLICA, "serverless", name)
	if info != "" {
		return errors.New("the function name is duplicated")
	}
	
	err := function.CreateImage(path, name)
	if err != nil {
		log.Error("[InitFunc] create image error: ", err)
		return err
	}
	ImageName := serverIp + ":5000/" + name + ":latest"
	replicaSet := GenerateReplicaSet(name, "serverless", ImageName, 0)

	// create the record
	log.Info("[InitFunc] create the record")
	autoscaler.RecordMutex.Lock()
	autoscaler.RecordMap[name] = &autoscaler.Record{
		Name:      name,
		Replicas:  0,
		PodIps:    make(map[string]int32),
		CallCount: 0,
	}
	autoscaler.RecordMutex.Unlock()
	log.Info("[InitFunc] create the record successfully")

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
	log.Info("[TriggerFunc] trigger function: ", name)
	retry := 3
	for retry > 0 {
		retry -= 1
		podIps, err := CheckPrepare(name)
		if err != nil {
			log.Error("[TriggerFunc] check prepare error: ", err)
			continue
		}

		// 2. load balance
		podIp, err := LoadBalance(name, podIps)
		if err != nil {
			log.Error("[TriggerFunc] load balance error: ", err)
			continue
		}

		// 3. trigger the function
		url := "http://" + podIp + ":8081/"
		var data interface{}
		err = json.Unmarshal(params, &data)
		prettyJSON, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Error("[TriggerFunc] marshal params error: ", err)
		}
		log.Info ("[TriggerFunc] prettyJSON: ", string(prettyJSON), "url: ", url)

		// 4. send the request
		// first check the connection
		err = CheckConnection(podIp)
		if err != nil {
			log.Error("[TriggerFunc] check connection error: ", err)
			continue
		}
		log.Info("[TriggerFunc] connection is finished")

		ret, err := utils.SendRequest("POST", params, url)
		log.Info("[TriggerFunc] ret: ", string(ret))
		result := bytes.NewBufferString(ret).Bytes()
		if err != nil {
			log.Error("[TriggerFunc] send request error: ", err)
			continue
		}

		return result, err
	}
	return nil, errors.New("trigger function error")
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
	log.Info("[DEleteFunc] delete record from record map")
	autoscaler.RecordMutex.Lock()
	autoscaler.DeleteRecord(name)
	autoscaler.RecordMutex.Unlock()

	log.Info("[DeleteFunc] delete record from record map")

	// 3. delete the image
	err = function.DeleteImage(name)
	if err != nil {
		log.Error("[DeleteFunc] delete image error: ", err)
		return err
	}

	// 4. ensure the replicaset is deleted

	return nil
}
