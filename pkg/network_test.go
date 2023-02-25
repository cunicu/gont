// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
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
	assert.NoError(t, err, "Failed to create network")
	defer n.Close()

	assert.Equal(t, n.Name, name, "Mismatching names")
	assert.True(t, hasNetwork(name))

	_, err = n.AddHost("h1")
	assert.NoError(t, err, "Failed to add host")

	_, err = netns.GetFromName(ns)
	assert.NoError(t, err, "Failed to get ns from name")
}

func TestNetworkNSPrefix(t *testing.T) {
	prefix := "pfx-"
	name := g.GenerateNetworkName()
	ns := fmt.Sprintf("%s%s-h1", prefix, name)

	n, err := g.NewNetwork(name, o.NSPrefix(prefix))
	assert.NoError(t, err, "Failed to create network: %s", err)
	defer n.Close()

	assert.Equal(t, n.Name, name, "Mismatching names")
	assert.True(t, hasNetwork(name))

	_, err = n.AddHost("h1")
	assert.NoError(t, err, "Failed to add host")

	_, err = netns.GetFromName(ns)
	assert.NoError(t, err, "Failed to get ns from name")
}

func TestNetworkGeneratedName(t *testing.T) {
	n1, err := g.NewNetwork("")
	assert.NoError(t, err, "Failed to create network")
	defer n1.Close()

	n2, err := g.NewNetwork("")
	assert.NoError(t, err, "Failed to create another network")
	defer n2.Close()
}

func TestNetworkExists(t *testing.T) {
	name := g.GenerateNetworkName()

	n1, err := g.NewNetwork(name)
	assert.NoError(t, err, "Failed to create network")
	defer n1.Close()

	n2, err := g.NewNetwork(name)
	if err == nil {
		defer n2.Close()
	}
	assert.Error(t, err, "Created network with existing name")
}
