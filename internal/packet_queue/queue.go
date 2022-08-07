package packet_prque

import (
	"container/heap"
	"sync"
	"time"

	"github.com/google/gopacket"
)

type Packet struct {
	CaptureInfo gopacket.CaptureInfo
	Data        []byte
}

type packetHeap []Packet

func (q packetHeap) Len() int {
	return len(q)
}

func (q packetHeap) Less(i, j int) bool {
	return q[i].CaptureInfo.Timestamp.Before(q[j].CaptureInfo.Timestamp)
}

func (q packetHeap) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *packetHeap) Push(x any) {
	*q = append(*q, x.(Packet))
}

func (q *packetHeap) Pop() any {
	old := *q
	n := len(old)
	x := old[n-1]
	*q = old[0 : n-1]
	return x
}

type PacketQueue struct {
	heap packetHeap
	lock sync.RWMutex
}

func New() *PacketQueue {
	return &PacketQueue{
		heap: []Packet{},
	}
}

func (q *PacketQueue) Push(pkt Packet) {
	q.lock.Lock()
	defer q.lock.Unlock()

	heap.Push(&q.heap, pkt)
}

func (q *PacketQueue) Pop() Packet {
	q.lock.Lock()
	defer q.lock.Unlock()

	pkt := heap.Pop(&q.heap).(Packet)

	return pkt
}

func (q *PacketQueue) Oldest() time.Time {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return q.heap[0].CaptureInfo.Timestamp
}

func (q *PacketQueue) Len() int {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return len(q.heap)
}
