package prque_test

import (
	"testing"
	"time"

	"github.com/stv0g/gont/internal/prque"
)

type item struct {
	ts time.Time
}

func (i item) Time() time.Time {
	return i.ts
}

func TestPriorityQueue(t *testing.T) {
	q := prque.New()

	it := func(t int) prque.Item {
		return item{
			ts: time.Unix(int64(t), 0),
		}
	}

	q.Push(it(4))
	q.Push(it(1))
	q.Push(it(2))
	q.Push(it(3))

	if q.Len() != 4 {
		t.Fail()
	}

	if it := q.Pop(); it.Time().Second() != 1 {
		t.Fail()
	}

	if it := q.Pop(); it.Time().Second() != 2 {
		t.Fail()
	}

	if o := q.Oldest(); o.Unix() != 3 {
		t.Fail()
	}

	if it := q.Pop(); it.Time().Second() != 3 {
		t.Fail()
	}

	if it := q.Pop(); it.Time().Second() != 4 {
		t.Fail()
	}

	if q.Len() != 0 {
		t.Fail()
	}
}
