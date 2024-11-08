// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"net"
	"testing"

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	"github.com/stretchr/testify/require"
)

func TestLinkAttrs(t *testing.T) {
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
			o.AddressIP("fc::1/64"),
			o.MTU(1000),
			o.Group(1234),
			o.AddressMACBytes([]byte{0, 0, 0, 0, 0, 1})),
		g.NewInterface("veth0", h2,
			o.AddressIP("10.0.0.2/24"),
			o.AddressIP("fc::2/64"),
			o.Group(5678),
			o.MTU(2000),
			o.AddressMACBytes([]byte{0, 0, 0, 0, 0, 2})))
	require.NoError(t, err, "Failed to setup link")

	h1Link, err := h1.NetlinkHandle().LinkByName("veth0")
	require.NoError(t, err, "Failed to get link details")

	require.Equal(t, h1Link.Attrs().MTU, 1000, "Mismatching MTU")
	require.Equal(t, h1Link.Attrs().Group, uint32(1234), "Mismatching device group")
	require.Equal(t, h1Link.Attrs().HardwareAddr, net.HardwareAddr([]byte{0, 0, 0, 0, 0, 1}), "Mismatching MAC address")

	h2Link, err := h2.NetlinkHandle().LinkByName("veth0")
	require.NoError(t, err, "Failed to get link details")

	require.Equal(t, h2Link.Attrs().MTU, 2000, "Mismatching MTU")
	require.Equal(t, h2Link.Attrs().Group, uint32(5678), "Mismatching device group")
	require.Equal(t, h2Link.Attrs().HardwareAddr, net.HardwareAddr([]byte{0, 0, 0, 0, 0, 2}), "Mismatching MAC address")
}
