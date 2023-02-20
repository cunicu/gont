// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"os"
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

// TestPingNAT performs and end-to-end ping test
// through a NAT topology
//
//	h1 <-> sw1 <-> nat1 <-> sw2 <-> h2
func TestPingNATIPv4(t *testing.T) {
	var (
		err      error
		n        *g.Network
		sw1, sw2 *g.Switch
		h1, h2   *g.Host
	)

	if n, err = g.NewNetwork(*nname, globalNetworkOptions...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
	defer n.Close()

	if sw1, err = n.AddSwitch("sw1"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	if sw2, err = n.AddSwitch("sw2"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	if h1, err = n.AddHost("h1",
		o.DefaultGatewayIP("10.0.1.1"),
		g.NewInterface("veth0", sw1,
			o.AddressIP("10.0.1.2/24")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if h2, err = n.AddHost("h2",
		o.DefaultGatewayIP("10.0.2.1"),
		g.NewInterface("veth0", sw2,
			o.AddressIP("10.0.2.2/24")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if _, err = n.AddNAT("nat1",
		g.NewInterface("veth0", sw1, o.SouthBound,
			o.AddressIP("10.0.1.1/24")),
		g.NewInterface("veth1", sw2, o.NorthBound,
			o.AddressIP("10.0.2.1/24")),
	); err != nil {
		t.Fatalf("Failed to create nat: %s", err)
	}

	if _, err = h1.Ping(h2); err != nil {
		t.Fatalf("Failed to ping h1 -> h2: %s", err)
	}

	if err = h1.Traceroute(h2); err != nil {
		t.Errorf("Failed to traceroute h1 -> h2: %s", err)
		t.Fail()
	}
}

// TestPingNATIPv6 performs and end-to-end ping test
// through a NAT topology using IPv6 addressing
//
//	h1 <-> sw1 <-> nat1 <-> sw2 <-> h2
func TestPingNATIPv6(t *testing.T) {
	var (
		err      error
		n        *g.Network
		sw1, sw2 *g.Switch
		h1, h2   *g.Host
	)

	if n, err = g.NewNetwork(*nname, globalNetworkOptions...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
	defer n.Close()

	if sw1, err = n.AddSwitch("sw1"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	if sw2, err = n.AddSwitch("sw2"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	if h1, err = n.AddHost("h1",
		o.DefaultGatewayIP("fc::1:1"),
		g.NewInterface("veth0", sw1,
			o.AddressIP("fc::1:2/112")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if h2, err = n.AddHost("h2",
		o.DefaultGatewayIP("fc::2:1"),
		g.NewInterface("veth0", sw2,
			o.AddressIP("fc::2:2/112")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if _, err = n.AddNAT("nat1",
		g.NewInterface("veth0", sw1, o.SouthBound,
			o.AddressIP("fc::1:1/112")),
		g.NewInterface("veth1", sw2, o.NorthBound,
			o.AddressIP("fc::2:1/112")),
	); err != nil {
		t.Fatalf("Failed to create nat: %s", err)
	}

	if _, err = h1.PingWithNetwork(h2, "ip6"); err != nil {
		t.Fatalf("Failed to ping h1 -> h2: %s", err)
	}

	if err = h1.Traceroute(h2, "-6"); err != nil {
		t.Errorf("Failed to traceroute h1 -> h2: %s", err)
		t.Fail()
	}
}

// TestPingDoubleNAT performs and end-to-end ping test
// through a double / carrier-grade NAT topology
//
//	h1 <-> sw1 <-> nat1 <-> sw2 <-> nat2 <-> sw3 <-> h2
func TestPingDoubleNAT(t *testing.T) {
	var (
		err           error
		n             *g.Network
		h1, h2        *g.Host
		sw1, sw2, sw3 *g.Switch
	)

	if n, err = g.NewNetwork(*nname, globalNetworkOptions...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
	defer n.Close()

	if sw1, err = n.AddSwitch("sw1"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	if sw2, err = n.AddSwitch("sw2"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	if sw3, err = n.AddSwitch("sw3"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	if h1, err = n.AddHost("h1",
		o.DefaultGatewayIP("10.0.1.1"),
		g.NewInterface("veth0", sw1,
			o.AddressIP("10.0.1.2/24")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if h2, err = n.AddHost("h2",
		o.DefaultGatewayIP("10.0.2.1"),
		g.NewInterface("veth0", sw3,
			o.AddressIP("10.0.2.2/24")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if _, err = n.AddNAT("nat1",
		o.DefaultGatewayIP("10.0.3.1"),
		g.NewInterface("veth1", sw1, o.SouthBound,
			o.AddressIP("10.0.1.1/24")),
		g.NewInterface("veth0", sw2, o.NorthBound,
			o.AddressIP("10.0.3.2/24")),
	); err != nil {
		t.Fatalf("Failed to create NAT router: %s", err)
	}

	if _, err = n.AddNAT("nat2",
		g.NewInterface("veth1", sw2, o.SouthBound,
			o.AddressIP("10.0.3.1/24")),
		g.NewInterface("veth0", sw3, o.NorthBound,
			o.AddressIP("10.0.2.1/24")),
	); err != nil {
		t.Fatalf("Failed to create NAT router: %s", err)
	}

	if _, err = h1.Ping(h2); err != nil {
		t.Fatalf("Failed to ping h1 <-> h2: %s", err)
	}

	if err = h1.Traceroute(h2); err != nil {
		t.Errorf("Failed to traceroute h1 -> h2: %s", err)
		t.Fail()
	}
}

// TestPingHostNAT performs and end-to-end ping test
// between a virtual host and the outside host network
//
//	h1 <-> sw <-> n1 (host) <-> external
func TestPingHostNAT(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		// GitHubs Azure based CI environment does not
		// allow to ping external targets
		t.Skip()
	}

	var (
		err error
		n   *g.Network
		sw1 *g.Switch
		h1  *g.Host
	)

	if n, err = g.NewNetwork(*nname, globalNetworkOptions...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
	defer n.Close()

	if sw1, err = n.AddSwitch("sw1"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	if h1, err = n.AddHost("h1",
		o.DefaultGatewayIP("10.0.0.1"),
		g.NewInterface("veth0", sw1,
			o.AddressIP("10.0.0.2/24")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if _, err := n.AddHostNAT("n1",
		g.NewInterface("veth0", sw1, o.SouthBound,
			o.AddressIP("10.0.0.1/24")),
	); err != nil {
		t.Fatalf("Failed to create host NAT: %s", err)
	}

	if _, err = h1.Run("ping", "-c", 1, "1.1.1.1"); err != nil {
		t.Fatalf("Failed to ping: %s", err)
	}

	if _, err = h1.Run("ping", "-c", 1, "www.rwth-aachen.de"); err != nil {
		t.Fail()
	}

	if _, err := h1.Ping(n.HostNode); err != nil {
		t.Fail()
	}
}
