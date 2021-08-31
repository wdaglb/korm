package korm

import (
	"container/list"
	"sync"
)

type Queue struct {
	data *list.List
}

var locLock sync.Mutex

func newQueue() *Queue {
	q := new(Queue)
	q.data = list.New()
	return q
}

func (q *Queue) push(v interface{}) {
	defer locLock.Unlock()
	locLock.Lock()
	q.data.PushFront(v)
}

func (q *Queue) pop() interface{} {
	defer locLock.Unlock()
	locLock.Lock()
	v := q.data.Front()
	val := v.Value
	q.data.Remove(v)
	return val
}

func (q *Queue) get() interface{} {
	defer locLock.Unlock()
	locLock.Lock()
	v := q.data.Front()
	val := v.Value
	return val
}
