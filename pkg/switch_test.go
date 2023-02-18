// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

// TestPing performs and end-to-end ping test
// between two hosts on a switched topology
//
//	h1 <-> sw1 <-> sw2 <-> h2
func TestPingCascadedSwitches(t *testing.T) {
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
		t.Fatalf("Failed to add switch: %s", err)
	}

	if sw2, err = n.AddSwitch("sw2"); err != nil {
		t.Fatalf("Failed to add switch: %s", err)
	}

	if h1, err = n.AddHost("h1",
		g.NewInterface("veth0", sw1,
			o.AddressIP("10.0.0.1/24")),
	); err != nil {
		t.Fatalf("Failed to add host: %s", err)
	}

	if h2, err = n.AddHost("h2",
		g.NewInterface("veth0", sw2,
			o.AddressIP("10.0.0.2/24")),
	); err != nil {
		t.Fatalf("Failed to add host: %s", err)
	}

	if err = n.AddLink(
		g.NewInterface("br-sw2", sw1),
		g.NewInterface("br-sw1", sw2),
	); err != nil {
		t.Fatalf("Failed to add link: %s", err)
	}

	if err := g.TestConnectivity(h1, h2); err != nil {
		t.Errorf("Failed to check connectivity: %s", err)
	}
}
