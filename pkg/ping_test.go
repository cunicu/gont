// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"testing"

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	"github.com/stretchr/testify/require"
)

// TestPingDualStack performs and end-to-end ping test
// between two hosts on a switched topology
// using both IPv4 and IPv6 addresses
//
//	h1 <-> sw <-> h2
func TestPingDualStack(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	sw, err := n.AddSwitch("sw")
	require.NoError(t, err, "Failed to create switch")

	h1, err := n.AddHost("h1",
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.0.1/24"),
			o.AddressIP("fc::1/64")))
	require.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2",
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.0.2/24"),
			o.AddressIP("fc::2/64")))
	require.NoError(t, err, "Failed to create host")

	for _, net := range []string{"ip4", "ip6"} {
		_, err = h1.PingWithNetwork(h2, net)
		require.NoError(t, err, "Failed to test connectivity")
	}
}

// TestPingIPv4 performs and end-to-end ping test
// between two hosts on a switched topology
// using IPv4 addressing
//
//	h1 <-> sw <-> h2
func TestPingIPv4(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	sw, err := n.AddSwitch("sw")
	require.NoError(t, err, "Failed to create switch")

	h1, err := n.AddHost("h1",
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.0.1/24")))
	require.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2",
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.0.2/24")))
	require.NoError(t, err, "Failed to create host")

	err = g.TestConnectivity(h1, h2)
	require.NoError(t, err, "Failed to test connectivity between hosts")
}

// TestPingIPv6 performs and end-to-end ping test
// between two hosts on a switched topology
// using IPv6 addressing
//
//	h1 <-> sw <-> h2
func TestPingIPv6(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	sw, err := n.AddSwitch("sw")
	require.NoError(t, err, "Failed to create switch")

	h1, err := n.AddHost("h1",
		g.NewInterface("veth0", sw,
			o.AddressIP("fc::1/64")))
	require.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2",
		g.NewInterface("veth0", sw,
			o.AddressIP("fc::2/64")))
	require.NoError(t, err, "Failed to create host")

	err = g.TestConnectivity(h1, h2)
	require.NoError(t, err, "Failed to test connectivity between hosts")
}

// TestPingDirect performs and end-to-end ping test between two
// directly connected hosts
//
// h1 <-> h2
func TestPingDirect(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h1, err := n.AddHost("h1")
	require.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2")
	require.NoError(t, err, "Failed to create host")

	err = n.AddLink(
		g.NewInterface("veth0", h1,
			o.AddressIP("10.0.0.1/24")),
		g.NewInterface("veth0", h2,
			o.AddressIP("10.0.0.2/24")))
	require.NoError(t, err, "Failed to connect hosts")

	_, err = h1.Run("cat", "/etc/hosts")
	require.NoError(t, err, "Failed to show /etc/hosts file")

	err = g.TestConnectivity(h1, h2)
	require.NoError(t, err, "Failed to test connectivity between hosts")
}

// TestPingMultiHop performs and end-to-end ping test
// through a routed multi-hop topology
//
//	h1 <-> sw1 <-> r1 <-> sw2 <-> h2
func TestPingMultiHop(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	sw1, err := n.AddSwitch("sw1")
	require.NoError(t, err, "Failed to create switch")

	sw2, err := n.AddSwitch("sw2")
	require.NoError(t, err, "Failed to create switch")

	h1, err := n.AddHost("h1",
		o.DefaultGatewayIP("10.0.1.1"),
		g.NewInterface("veth0", sw1,
			o.AddressIP("10.0.1.2/24"),
		))
	require.NoError(t, err, "Failed to add host")

	h2, err := n.AddHost("h2",
		o.DefaultGatewayIP("10.0.2.1"),
		g.NewInterface("veth0", sw2,
			o.AddressIP("10.0.2.2/24")))
	require.NoError(t, err, "Failed to add host")

	_, err = n.AddRouter("r1",
		g.NewInterface("veth0", sw1,
			o.AddressIP("10.0.1.1/24")),
		g.NewInterface("veth1", sw2,
			o.AddressIP("10.0.2.1/24")))
	require.NoError(t, err, "Failed to add router")

	err = g.TestConnectivity(h1, h2)
	require.NoError(t, err, "Failed to test connectivity between hosts")

	err = h1.Traceroute(h2)
	require.NoError(t, err, "Failed to traceroute from h1 to h2")
}

func TestPingLoopback(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h, err := n.AddHost("h")
	require.NoError(t, err, "Failed to add host")

	_, err = h.Run("ping", "-4", "-c", 1, "localhost")
	require.NoError(t, err, "Failed to ping")

	_, err = h.Run("ping", "-6", "-c", 1, "localhost")
	require.NoError(t, err, "Failed to ping")
}

func TestPingSelf(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h1, err := n.AddHost("h1")
	require.NoError(t, err, "Failed to add host")

	h2, err := n.AddHost("h2")
	require.NoError(t, err, "Failed to add host")

	err = n.AddLink(
		g.NewInterface("veth0", h1,
			o.AddressIP("10.0.0.1/24"),
			o.AddressIP("fc::1/64")),
		g.NewInterface("veth0", h2,
			o.AddressIP("10.0.0.2/24"),
			o.AddressIP("fc::2/64")))
	require.NoError(t, err, "Failed to add link")

	_, err = h1.PingWithNetwork(h1, "ip")
	require.NoError(t, err, "Failed to ping")

	_, err = h1.PingWithNetwork(h1, "ip6")
	require.NoError(t, err, "Failed to ping")
}
