package watch

import (
	"container/list"
	"sync"
)

type ThreadSafeList struct {
	mux  sync.Mutex
	List *list.List
}

func NewThreadSafeList() *ThreadSafeList {
	return &ThreadSafeList{List: list.New()}
}

func (l *ThreadSafeList) Len() int {
	l.mux.Lock()
	defer l.mux.Unlock()

	return l.List.Len()
}

func (l *ThreadSafeList) Front() *list.Element {
	l.mux.Lock()
	defer l.mux.Unlock()

	return l.List.Front()
}

func (l *ThreadSafeList) Back() *list.Element {
	l.mux.Lock()
	defer l.mux.Unlock()

	return l.List.Back()
}

func (l *ThreadSafeList) PushFront(v interface{}) *list.Element {
	l.mux.Lock()
	defer l.mux.Unlock()

	return l.List.PushFront(v)
}

func (l *ThreadSafeList) PushBack(v interface{}) *list.Element {
	l.mux.Lock()
	defer l.mux.Unlock()

	return l.List.PushBack(v)
}

func (l *ThreadSafeList) Remove(e *list.Element) interface{} {
	l.mux.Lock()
	defer l.mux.Unlock()

	return l.List.Remove(e)
}
