package utils

import (
	"fmt"
	"sync"
)

var locks = sync.Map{}
var lockForLocks = sync.Mutex{}

func RLock(namespace, podName string) {
	key := fmt.Sprintf("%s-%s", namespace, podName)
	if mutex, ok := locks.Load(key); ok {
		mutex.(*sync.RWMutex).RLock()
	} else {
		lockForLocks.Lock()
		mutex := &sync.RWMutex{}
		mutex.RLock()
		locks.Store(key, mutex)
		lockForLocks.Unlock()
	}
}

func Lock(namespace, podName string) {
	key := fmt.Sprintf("%s-%s", namespace, podName)
	if mutex, ok := locks.Load(key); ok {
		mutex.(*sync.RWMutex).Lock()
	} else {
		lockForLocks.Lock()
		mutex := &sync.RWMutex{}
		mutex.Lock()
		locks.Store(key, mutex)
		lockForLocks.Unlock()
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
