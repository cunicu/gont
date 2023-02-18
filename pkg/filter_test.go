// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"net"
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	fo "github.com/stv0g/gont/pkg/options/filters"
	"golang.org/x/sys/unix"
)

func TestFilterIPv4(t *testing.T) {
	var (
		err        error
		n          *g.Network
		sw         *g.Switch
		h1, h2, h3 *g.Host
	)

	if n, err = g.NewNetwork(*nname, globalNetworkOptions...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
	defer n.Close()

	if sw, err = n.AddSwitch("sw"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	_, flt, err := net.ParseCIDR("10.0.3.0/24")
	if err != nil {
		t.Fail()
	}

	if h1, err = n.AddHost("h1",
		o.Filter(g.FilterInput,
			fo.Protocol(unix.AF_INET),
			fo.TransportProtocol(unix.IPPROTO_ICMP),
			fo.Source(flt),
			fo.Drop,
		),
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.1.1/16")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if h2, err = n.AddHost("h2",
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.2.1/16")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if h3, err = n.AddHost("h3",
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.3.1/16")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if _, err := h1.Ping(h2); err != nil {
		t.Fail()
	}

	if _, err := h1.Ping(h3); err == nil {
		t.Fail()
	}
}

func TestFilterIPv6(t *testing.T) {
	var (
		err        error
		n          *g.Network
		sw         *g.Switch
		h1, h2, h3 *g.Host
	)

	if n, err = g.NewNetwork(*nname, globalNetworkOptions...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
	defer n.Close()

	if sw, err = n.AddSwitch("sw"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	_, flt, err := net.ParseCIDR("fc00:0:0:3::1/64")
	if err != nil {
		t.Fail()
	}

	if h1, err = n.AddHost("h1",
		o.Filter(g.FilterInput,
			fo.Protocol(unix.AF_INET6),
			fo.TransportProtocol(unix.IPPROTO_ICMPV6),
			fo.Source(flt),
			fo.Drop,
		),
		g.NewInterface("veth0", sw,
			o.AddressIP("fc00:0:0:1::1/56")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if h2, err = n.AddHost("h2",
		g.NewInterface("veth0", sw,
			o.AddressIP("fc00:0:0:2::1/56")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if h3, err = n.AddHost("h3",
		g.NewInterface("veth0", sw,
			o.AddressIP("fc00:0:0:3::1/56")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if _, err := h1.Ping(h2); err != nil {
		t.Fail()
	}

	if _, err := h1.Ping(h3); err == nil {
		t.Fail()
	}
}
