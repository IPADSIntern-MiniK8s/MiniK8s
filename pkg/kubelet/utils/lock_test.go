package utils

import (
	"sync"
	"testing"
)

func TestReadLock(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(1)

	RLock("test", "pod1")
	go func() {
		RLock("test", "pod1")
		wg.Done()
	}()
	wg.Wait()
	RUnLock("test", "pod1")
}

//deadlock is correct

//func TestWriteLock(t *testing.T) {
//	wg := sync.WaitGroup{}
//	wg.Add(1)
//
//	RLock("test", "pod1")
//	go func() {
//		Lock("test", "pod1")
//		wg.Done()
//	}()
//	wg.Wait()
//	RUnLock("test", "pod1")
//}
