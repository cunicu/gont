// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"fmt"
	"testing"

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	"github.com/stretchr/testify/require"
)

func prepareNetwork(t *testing.T, i int) *g.Network {
	pfx := fmt.Sprintf("net%d-", i)

	address := func(j int) o.Address {
		return o.AddressIP("fc::%d:%d/112", i, j)
	}

	n, err := g.NewNetwork("", globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")

	sw, err := n.AddSwitch(pfx + "sw")
	require.NoError(t, err, "Failed to create switch")

	_, err = n.AddHost(pfx+"h1",
		g.NewInterface("veth0", sw,
			address(1)))
	require.NoError(t, err, "Failed to create host")

	_, err = n.AddHost(pfx+"h2",
		g.NewInterface("veth0", sw,
			address(2)))
	require.NoError(t, err, "Failed to create host")

	return n
}

func TestMultipleNetworks(t *testing.T) {
	n1 := prepareNetwork(t, 1)
	defer n1.Close()

	n2 := prepareNetwork(t, 2)
	defer n2.Close()

	// Connectivity within the network must succeed
	err := g.TestConnectivity(n1.Hosts()...)
	require.NoError(t, err, "Connectivity tests between hosts on same network must succeed")

	// Connectivity within the network must succeed
	err = g.TestConnectivity(n2.Hosts()...)
	require.NoError(t, err, "Connectivity tests between hosts on same network must succeed")

	// Connectivity between the networks must fail
	all := append(n1.Hosts(), n2.Hosts()...)
	err = g.TestConnectivity(all...)
	require.Error(t, err, "Connectivity tests between hosts on different networks must fail")
}
