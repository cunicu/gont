package packet_prque_test

import (
	"testing"
	"time"

	"github.com/google/gopacket"
	pprque "github.com/stv0g/gont/internal/packet_queue"
)

func TestPacketPriorityQueue(t *testing.T) {
	q := pprque.New()

	pkt := func(t int64) pprque.Packet {
		return pprque.Packet{
			CaptureInfo: gopacket.CaptureInfo{Timestamp: time.Unix(t, 0)},
		}
	}

	q.Push(pkt(4))
	q.Push(pkt(1))
	q.Push(pkt(2))
	q.Push(pkt(3))

	if q.Len() != 4 {
		t.Fail()
	}

	if pkt := q.Pop(); pkt.CaptureInfo.Timestamp.Unix() != 1 {
		t.Fail()
	}

	if pkt := q.Pop(); pkt.CaptureInfo.Timestamp.Unix() != 2 {
		t.Fail()
	}

	if o := q.Oldest(); o.Unix() != 3 {
		t.Fail()
	}

	if pkt := q.Pop(); pkt.CaptureInfo.Timestamp.Unix() != 3 {
		t.Fail()
	}

	if pkt := q.Pop(); pkt.CaptureInfo.Timestamp.Unix() != 4 {
		t.Fail()
	}

	if q.Len() != 0 {
		t.Fail()
	}
}
