// SPDX-FileCopyrightText: 2023 Marin Atanasov Nikolov <dnaeon@gmail.com>
// SPDX-License-Identifier: BSD-2-Clause

package prque_test

import (
	"testing"

	"cunicu.li/gont/v2/internal/prque"
)

func TestPriorityQueue(t *testing.T) {
	queue := prque.New[string, int64]()
	queue.Put("apple", 10)
	queue.Put("banana", 3)
	queue.Put("pear", 20)
	queue.Put("orange", 15)

	want := []struct {
		value    string
		priority int64
	}{
		{"banana", 3},
		{"apple", 10},
		{"orange", 15},
		{"pear", 20},
	}

	i := 0
	for !queue.IsEmpty() {
		val, prio := queue.Get()
		if val != want[i].value || prio != want[i].priority {
			t.Fatalf("want %q with priority %d, got %q with priority %d", want[i].value, want[i].priority, val, prio)
		}
		i++
	}
}
