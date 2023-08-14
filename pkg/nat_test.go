// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"os"
	"testing"

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	"github.com/stretchr/testify/require"
)

// TestPingNAT performs and end-to-end ping test
// through a NAT topology
//
//	h1 <-> sw1 <-> nat1 <-> sw2 <-> h2
func TestPingNATIPv4(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.Close()

	sw1, err := n.AddSwitch("sw1")
	require.NoError(t, err, "Failed to create switch")

	sw2, err := n.AddSwitch("sw2")
	require.NoError(t, err, "Failed to create switch")

	h1, err := n.AddHost("h1",
		o.DefaultGatewayIP("10.0.1.1"),
		g.NewInterface("veth0", sw1,
			o.AddressIP("10.0.1.2/24")))
	require.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2",
		o.DefaultGatewayIP("10.0.2.1"),
		g.NewInterface("veth0", sw2,
			o.AddressIP("10.0.2.2/24")))
	require.NoError(t, err, "Failed to create host")

	_, err = n.AddNAT("nat1",
		g.NewInterface("veth0", sw1, o.SouthBound,
			o.AddressIP("10.0.1.1/24")),
		g.NewInterface("veth1", sw2, o.NorthBound,
			o.AddressIP("10.0.2.1/24")))
	require.NoError(t, err, "Failed to create NAT")

	_, err = h1.Ping(h2)
	require.NoError(t, err, "Failed to ping h1 -> h2")

	err = h1.Traceroute(h2)
	require.NoError(t, err, "Failed to traceroute h1 -> h2")
}

// TestPingNATIPv6 performs and end-to-end ping test
// through a NAT topology using IPv6 addressing
//
//	h1 <-> sw1 <-> nat1 <-> sw2 <-> h2
func TestPingNATIPv6(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.Close()

	sw1, err := n.AddSwitch("sw1")
	require.NoError(t, err, "Failed to create switch")

	sw2, err := n.AddSwitch("sw2")
	require.NoError(t, err, "Failed to create switch")

	h1, err := n.AddHost("h1",
		o.DefaultGatewayIP("fc::1:1"),
		g.NewInterface("veth0", sw1,
			o.AddressIP("fc::1:2/112")))
	require.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2",
		o.DefaultGatewayIP("fc::2:1"),
		g.NewInterface("veth0", sw2,
			o.AddressIP("fc::2:2/112")))
	require.NoError(t, err, "Failed to create host")

	_, err = n.AddNAT("nat1",
		g.NewInterface("veth0", sw1, o.SouthBound,
			o.AddressIP("fc::1:1/112")),
		g.NewInterface("veth1", sw2, o.NorthBound,
			o.AddressIP("fc::2:1/112")))
	require.NoError(t, err, "Failed to create NAT")

	_, err = h1.PingWithNetwork(h2, "ip6")
	require.NoError(t, err, "Failed to ping h1 -> h2")

	err = h1.Traceroute(h2, "-6")
	require.NoError(t, err, "Failed to traceroute h1 -> h2")
}

// TestPingDoubleNAT performs and end-to-end ping test
// through a double / carrier-grade NAT topology
//
//	h1 <-> sw1 <-> nat1 <-> sw2 <-> nat2 <-> sw3 <-> h2
func TestPingDoubleNAT(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.Close()

	sw1, err := n.AddSwitch("sw1")
	require.NoError(t, err, "Failed to create switch")

	sw2, err := n.AddSwitch("sw2")
	require.NoError(t, err, "Failed to create switch")

	sw3, err := n.AddSwitch("sw3")
	require.NoError(t, err, "Failed to create switch")

	h1, err := n.AddHost("h1",
		o.DefaultGatewayIP("10.0.1.1"),
		g.NewInterface("veth0", sw1,
			o.AddressIP("10.0.1.2/24")))
	require.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2",
		o.DefaultGatewayIP("10.0.2.1"),
		g.NewInterface("veth0", sw3,
			o.AddressIP("10.0.2.2/24")))
	require.NoError(t, err, "Failed to create host")

	_, err = n.AddNAT("nat1",
		o.DefaultGatewayIP("10.0.3.1"),
		g.NewInterface("veth1", sw1, o.SouthBound,
			o.AddressIP("10.0.1.1/24")),
		g.NewInterface("veth0", sw2, o.NorthBound,
			o.AddressIP("10.0.3.2/24")))
	require.NoError(t, err, "Failed to create NAT router")

	_, err = n.AddNAT("nat2",
		g.NewInterface("veth1", sw2, o.SouthBound,
			o.AddressIP("10.0.3.1/24")),
		g.NewInterface("veth0", sw3, o.NorthBound,
			o.AddressIP("10.0.2.1/24")))
	require.NoError(t, err, "Failed to create NAT router")

	_, err = h1.Ping(h2)
	require.NoError(t, err, "Failed to ping h1 <-> h2")

	err = h1.Traceroute(h2)
	require.NoError(t, err, "Failed to traceroute h1 -> h2")
}

// TestPingHostNAT performs and end-to-end ping test
// between a virtual host and the outside host network
//
//	h1 <-> sw <-> n1 (host) <-> external
func TestPingHostNAT(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		t.Skip("GitHubs Azure based CI environment does not allow to ping external targets")
	}

	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.Close()

	sw1, err := n.AddSwitch("sw1")
	require.NoError(t, err, "Failed to create switch")

	h1, err := n.AddHost("h1",
		o.DefaultGatewayIP("10.0.0.1"),
		g.NewInterface("veth0", sw1,
			o.AddressIP("10.0.0.2/24")))
	require.NoError(t, err, "Failed to create host")

	_, err = n.AddHostNAT("n1",
		g.NewInterface("veth0", sw1, o.SouthBound,
			o.AddressIP("10.0.0.1/24")))
	require.NoError(t, err, "Failed to create host NAT")

	_, err = h1.Run("ping", "-c", 1, "1.1.1.1")
	require.NoError(t, err, "Failed to ping")

	_, err = h1.Run("ping", "-c", 1, "www.rwth-aachen.de")
	require.NoError(t, err)

	_, err = h1.Ping(n.HostNode)
	require.NoError(t, err)
}
