package utils

import (
	"fmt"
	"sync"
)

var locks = sync.Map{}

func RLock(namespace, podName string) {
	key := fmt.Sprintf("%s-%s", namespace, podName)
	if mutex, ok := locks.Load(key); ok {
		mutex.(*sync.RWMutex).RLock()
	} else {
		mutex := &sync.RWMutex{}
		mutex.RLock()
		locks.Store(key, mutex)
	}
}

func Lock(namespace, podName string) {
	key := fmt.Sprintf("%s-%s", namespace, podName)
	if mutex, ok := locks.Load(key); ok {
		mutex.(*sync.RWMutex).Lock()
	} else {
		mutex := &sync.RWMutex{}
		mutex.Lock()
		locks.Store(key, mutex)
	}
}
func RUnLock(namespace, podName string) {
	key := fmt.Sprintf("%s-%s", namespace, podName)
	if mutex, ok := locks.Load(key); ok {
		mutex.(*sync.RWMutex).RUnlock()
	} else {
		panic("unlock a non-exist lock")
	}
}

func UnLock(namespace, podName string) {
	key := fmt.Sprintf("%s-%s", namespace, podName)
	if mutex, ok := locks.Load(key); ok {
		mutex.(*sync.RWMutex).Unlock()
	} else {
		panic("unlock a non-exist lock")
	}
}
