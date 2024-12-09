// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"fmt"
	"testing"

	g "cunicu.li/gont/v2/pkg"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netns"
)

func hasNetwork(name string) bool {
	for _, n := range g.NetworkNames() {
		if n == name {
			return true
		}
	}
	return false
}

func TestNamedNetwork(t *testing.T) {
	name := g.GenerateNetworkName()
	ns := fmt.Sprintf("gont-%s-h1", name)

	n, err := g.NewNetwork(name)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	require.Equal(t, n.Name, name, "Mismatching names")
	require.True(t, hasNetwork(name))

	_, err = n.AddHost("h1")
	require.NoError(t, err, "Failed to add host")

	_, err = netns.GetFromName(ns)
	require.NoError(t, err, "Failed to get ns from name")
}

func TestNetworkGeneratedName(t *testing.T) {
	n1, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n1.Close()

	n2, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create another network")
	defer n2.Close()
}

func TestNetworkExists(t *testing.T) {
	name := g.GenerateNetworkName()

	n1, err := g.NewNetwork(name)
	require.NoError(t, err, "Failed to create network")
	defer n1.Close()

	n2, err := g.NewNetwork(name)
	if err == nil {
		defer n2.Close()
	}
	require.Error(t, err, "Created network with existing name")
}
