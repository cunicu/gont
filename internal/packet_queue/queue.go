package packet_prque

import (
	"container/heap"
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
	packetHeap
}

func New() *PacketQueue {
	return &PacketQueue{
		packetHeap: []Packet{},
	}
}

func (q *PacketQueue) Push(pkt Packet) {
	heap.Push(&q.packetHeap, pkt)
}

func (q *PacketQueue) Pop() Packet {
	pkt := heap.Pop(&q.packetHeap).(Packet)

	return pkt
}

func (q *PacketQueue) Oldest() time.Time {
	return q.packetHeap[0].CaptureInfo.Timestamp
}
