// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"testing"

	g "github.com/stv0g/gont/pkg"
	"github.com/vishvananda/netns"
)

func TestNamespace(t *testing.T) {
	nsName := "gont-testing-ns"

	// delete stale namespaces from previous runs
	netns.DeleteNamed(nsName)

	n, err := g.NewNamespace(nsName)
	if err != nil {
		t.Fatalf("Failed to create new namespace: %s", err)
	}
	defer n.Close()

	if err := n.RunFunc(func() error {
		nsh, err := netns.Get()
		if err != nil {
			return err
		}
		if !nsh.Equal(n.NsHandle) {
			t.Errorf("NShandle mismatch: %v != %v", nsh, n.NsHandle)
		}
		return nil
	}); err != nil {
		t.Errorf("Failed to run func: %s", err)
	}
}
