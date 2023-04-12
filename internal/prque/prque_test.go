// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package prque_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stv0g/gont/v2/internal/prque"
)

type item struct {
	ts time.Time
}

func (i item) Time() time.Time {
	return i.ts
}

func TestPriorityQueue(t *testing.T) {
	q := prque.New()

	itf := func(t int) prque.Item {
		return item{
			ts: time.Unix(int64(t), 0),
		}
	}

	q.Push(itf(4))
	q.Push(itf(1))
	q.Push(itf(2))
	q.Push(itf(3))

	require.Equal(t, q.Len(), 4)

	it := q.Pop()
	require.Equal(t, it.Time().Second(), 1)

	it = q.Pop()
	require.Equal(t, it.Time().Second(), 2)

	o := q.Oldest()
	require.EqualValues(t, o.Unix(), 3)

	it = q.Pop()
	require.Equal(t, it.Time().Second(), 3)

	it = q.Pop()
	require.Equal(t, it.Time().Second(), 4)

	require.Zero(t, q.Len())
}
