// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	g "github.com/stv0g/gont/pkg"
	"github.com/vishvananda/netns"
)

func TestNamespace(t *testing.T) {
	nsName := "gont-testing-ns"

	// delete stale namespaces from previous runs
	netns.DeleteNamed(nsName) //nolint:errcheck

	n, err := g.NewNamespace(nsName)
	assert.NoError(t, err, "Failed to create new namespace")
	defer n.Close()

	err = n.RunFunc(func() error {
		nsh, err := netns.Get()
		if err != nil {
			return err
		}
		assert.True(t, nsh.Equal(n.NsHandle), "NShandle mismatch: %v != %v", nsh, n.NsHandle)
		return nil
	})
	assert.NoError(t, err, "Failed to run func")
}
