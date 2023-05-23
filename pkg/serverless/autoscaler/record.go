package autoscaler

import "sync"

type Record struct {
	// Name is the name of the function
	Name string `json:"name"`
	// the current replica number of the function
	Replicas int32 `json:"replicas"`
	// the podIps that the function has deployed on
	PodIps map[string]int32 `json:"podIps"`
	// the call count of the function
	CallCount int32 `json:"callCount"`
}



var (
	RecordMap = make(map[string]*Record)
	RecordMutex sync.RWMutex	// protect the access of RecordMap
)


func GetRecord(name string) *Record {
	return RecordMap[name]
}

func SetRecord(name string, record *Record) {
	RecordMap[name] = record
}

func DeleteRecord(name string) {
	delete(RecordMap, name)
}


func UpdateRecord(name string) {
	record := GetRecord(name)
	if record == nil {
		return 
	}
	record.CallCount++
	SetRecord(name, record)
}


