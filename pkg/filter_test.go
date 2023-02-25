// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	fo "github.com/stv0g/gont/pkg/options/filters"
	"golang.org/x/sys/unix"
)

func TestFilterIPv4(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	assert.NoError(t, err, "Failed to create network")
	defer n.Close()

	sw, err := n.AddSwitch("sw")
	assert.NoError(t, err, "Failed to create switch")

	_, flt, err := net.ParseCIDR("10.0.3.0/24")
	assert.NoError(t, err, "Failed to parse CIDR")

	h1, err := n.AddHost("h1",
		o.Filter(g.FilterInput,
			fo.Protocol(unix.AF_INET),
			fo.TransportProtocol(unix.IPPROTO_ICMP),
			fo.Source(flt),
			fo.Drop),
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.1.1/16")))
	assert.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2",
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.2.1/16")))
	assert.NoError(t, err, "Failed to create host")

	h3, err := n.AddHost("h3",
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.3.1/16")))
	assert.NoError(t, err, "Failed to create host")

	_, err = h1.Ping(h2)
	assert.NoError(t, err, "Failed to ping h2")

	_, err = h1.Ping(h3)
	assert.Error(t, err, "Succeeded to ping h1")
}

func TestFilterIPv6(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	assert.NoError(t, err, "Failed to create network")
	defer n.Close()

	sw, err := n.AddSwitch("sw")
	assert.NoError(t, err, "Failed to create switch")

	_, flt, err := net.ParseCIDR("fc00:0:0:3::1/64")
	assert.NoError(t, err, "Failed to parse CIDR")

	h1, err := n.AddHost("h1",
		o.Filter(g.FilterInput,
			fo.Protocol(unix.AF_INET6),
			fo.TransportProtocol(unix.IPPROTO_ICMPV6),
			fo.Source(flt),
			fo.Drop),
		g.NewInterface("veth0", sw,
			o.AddressIP("fc00:0:0:1::1/56")))
	assert.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2",
		g.NewInterface("veth0", sw,
			o.AddressIP("fc00:0:0:2::1/56")))
	assert.NoError(t, err, "Failed to create host")

	h3, err := n.AddHost("h3",
		g.NewInterface("veth0", sw,
			o.AddressIP("fc00:0:0:3::1/56")))
	assert.NoError(t, err, "Failed to create host")

	_, err = h1.Ping(h2)
	assert.NoError(t, err, "Failed to ping h2")

	_, err = h1.Ping(h3)
	assert.Error(t, err, "Succeeded to ping h3")
}
