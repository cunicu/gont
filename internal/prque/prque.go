// SPDX-FileCopyrightText: 2023 Marin Atanasov Nikolov <dnaeon@gmail.com>
// SPDX-License-Identifier: BSD-2-Clause

package prque

import (
	"cmp"
	"container/heap"
	"sync"
)

// Item represents an item from the priority queue.
type Item[T any, V cmp.Ordered] struct {
	// The value associated with the item
	Value T

	// The priority of the item
	Priority V
}

// PriorityQueue is a priority queue implementation based
// container/heap.
type PriorityQueue[T any, V cmp.Ordered] struct {
	sync.Mutex
	items []*Item[T, V]
}

// New creates a new priority queue, containing items of type T with
// priority V.
func New[T any, V cmp.Ordered]() *PriorityQueue[T, V] {
	pq := &PriorityQueue[T, V]{
		items: make([]*Item[T, V], 0),
	}
	heap.Init(pq)

	return pq
}

// Len implements sort.Interface.
func (pq *PriorityQueue[T, V]) Len() int {
	return len(pq.items)
}

// Less implements sort.Interface.
func (pq *PriorityQueue[T, V]) Less(i, j int) bool {
	return pq.items[i].Priority < pq.items[j].Priority
}

// Swap implements sort.Interface.
func (pq *PriorityQueue[T, V]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// Push implements heap.Interface.
func (pq *PriorityQueue[T, V]) Push(x any) {
	item := x.(*Item[T, V]) //nolint:forcetypeassert
	pq.items = append(pq.items, item)
}

// Pop implements heap.Interface.
func (pq *PriorityQueue[T, V]) Pop() any {
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // Avoid memory leak
	pq.items = old[0 : n-1]

	return item
}

// Put adds a value with the given priority to the priority queue.
func (pq *PriorityQueue[T, V]) Put(value T, priority V) {
	pq.Lock()
	defer pq.Unlock()

	heap.Push(pq, &Item[T, V]{
		Value:    value,
		Priority: priority,
	})
}

// Get returns the next item from the priority queue
func (pq *PriorityQueue[T, V]) Get() (T, V) {
	pq.Lock()
	defer pq.Unlock()

	item := heap.Pop(pq).(*Item[T, V]) //nolint:forcetypeassert

	return item.Value, item.Priority
}

// Peek returns the next time from the priority queue without dequeing it.
func (pq *PriorityQueue[T, V]) Peek() (T, V) {
	pq.Lock()
	defer pq.Unlock()

	item := pq.items[0]

	return item.Value, item.Priority
}

// IsEmpty returns a boolean indicating whether the priority queue is
// empty or not.
func (pq *PriorityQueue[T, V]) IsEmpty() bool {
	pq.Lock()
	defer pq.Unlock()

	return pq.Len() == 0
}
