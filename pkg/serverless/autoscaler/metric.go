package autoscaler

import (
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/apiobject"
	"minik8s/config"
	"minik8s/utils"
	"time"
)

// query the pod ips of the replicaSet
func QueryPodIps() (map[string][]string, error) {
	// find all the pods of the replicaSet
	podUrl := "http://" + config.ApiServerIp + "/api/v1/namespaces/serverless/pods"
	response, err := utils.SendRequest("GET", nil, podUrl)
	if err != nil {
		log.Error("[QueryPodIps] error getting pods: ", err)
		return nil, err
	}

	result := make(map[string][]string)
	podTool := &apiobject.Pod{}
	podList, error := podTool.UnMarshalJsonList([]byte(response))
	if error != nil {
		log.Error("[QueryPodIps] error unmarshalling pods: ", error)
		return nil, error
	}

	for _, pod := range podList {
		pos, ok := result[pod.Data.Name]
		if !ok {
			pos = make([]string, 0)
		} 
		pos = append(pos, pod.Status.PodIp)
		result[pod.Data.Name] = pos
	}
	return result, nil
}


// PeriodicMetric check the invoke frequency periodically, delete the function if it is not invoked for a long time
func PeriodicMetric(timeInterval int) {
	for {
		// get all replicas
		replicaUrl := "http://" + config.ApiServerIp + "/api/v1/namespaces/serverless/replicas"
		response, err := utils.SendRequest("GET", nil, replicaUrl)
		if err != nil {
			log.Error("[PeriodicMetric] error getting replicas: ", err)
			continue
		}

		var replicaTool = &apiobject.ReplicationController{}
		replicaList, err := replicaTool.UnMarshalJSONList([]byte(response))
		if err != nil {
			log.Error("[PeriodicMetric] error unmarshalling replicas: ", err)
			continue
		}

		// get all replicaSet's pod ips
		if err != nil {
			log.Error("[PeriodicMetric] error querying pod ips: ", err)
			continue
		}

		// update the replicas information
		for _, replica := range replicaList {
			// get the according record in map
			RecordMutex.RLock()
			record := GetRecord(replica.Data.Name)
			RecordMutex.RUnlock()
			if record == nil {
				record = &Record{
					Name:      replica.Data.Name,
					Replicas:  replica.Status.Replicas,
					CallCount: 0,
				}
				RecordMutex.Lock()
				SetRecord(replica.Data.Name, record)
				RecordMutex.Unlock()
			} else {
				// if the call times is 0, scale to zero
				// scale according to the call times
				replica.Status.Scale = record.CallCount
				
				// update the replicaset
				if replica.Status.Scale != replica.Status.Replicas {
					replicaUrl := "http://" + config.ApiServerIp + "/api/v1/namespaces/serverless/replicas/" + replica.Data.Name + "/update"
					replicaJson, err := replica.MarshalJSON()
					if err != nil {
						log.Error("[PeriodicMetric] error marshalling replicas: ", err)
						continue
					}
					_, err = utils.SendRequest("PUT", replicaJson, replicaUrl)
					if err != nil {
						log.Error("[PeriodicMetric] error updating replicas: ", err)
						continue
					}
				}

				record.Replicas = replica.Status.Replicas
				record.CallCount = 0

				RecordMutex.Lock()
				SetRecord(replica.Data.Name, record)
				RecordMutex.Unlock()
			}	
		}

		time.Sleep(time.Duration(timeInterval) * time.Second)
	}	
}
	