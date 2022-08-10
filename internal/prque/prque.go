package prque

import (
	"container/heap"
	"sync"
	"time"
)

type Item interface {
	Time() time.Time
}

type heapl []Item

func (q heapl) Len() int {
	return len(q)
}

func (q heapl) Less(i, j int) bool {
	it := q[i].Time()
	jt := q[j].Time()

	return it.Before(jt)
}

func (q heapl) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *heapl) Push(x any) {
	*q = append(*q, x.(Item))
}

func (q *heapl) Pop() any {
	old := *q
	n := len(old)
	x := old[n-1]
	*q = old[0 : n-1]
	return x
}

type PriorityQueue struct {
	heap heapl
	lock sync.RWMutex
}

func New() *PriorityQueue {
	return &PriorityQueue{
		heap: []Item{},
	}
}

func (q *PriorityQueue) Push(item Item) {
	q.lock.Lock()
	defer q.lock.Unlock()

	heap.Push(&q.heap, item)
}

func (q *PriorityQueue) Pop() Item {
	q.lock.Lock()
	defer q.lock.Unlock()

	item := heap.Pop(&q.heap).(Item)

	return item
}

func (q *PriorityQueue) Oldest() time.Time {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return q.heap[0].Time()
}

func (q *PriorityQueue) Len() int {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return len(q.heap)
}
