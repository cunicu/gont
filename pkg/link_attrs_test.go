// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"bytes"
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

func TestLinkAttrs(t *testing.T) {
	var (
		err    error
		n      *g.Network
		h1, h2 *g.Host
	)

	if n, err = g.NewNetwork(*nname, globalNetworkOptions...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
	defer n.Close()

	if h1, err = n.AddHost("h1"); err != nil {
		t.Fatalf("Failed to add host: %s", err)
	}

	if h2, err = n.AddHost("h2"); err != nil {
		t.Fatalf("Failed to add host: %s", err)
	}

	if err := n.AddLink(
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
			o.AddressMACBytes([]byte{0, 0, 0, 0, 0, 2})),
	); err != nil {
		t.Errorf("Failed to setup link: %s", err)
	}

	h1Link, err := h1.NetlinkHandle().LinkByName("veth0")
	if err != nil {
		t.Errorf("Failed to get link details: %s", err)
	}

	if h1Link.Attrs().MTU != 1000 {
		t.Errorf("Mismatching MTU")
	}

	if h1Link.Attrs().Group != 1234 {
		t.Errorf("Mismatching device group")
	}

	if !bytes.Equal(h1Link.Attrs().HardwareAddr, []byte{0, 0, 0, 0, 0, 1}) {
		t.Errorf("Mismatching MAC address")
	}

	h2Link, err := h2.NetlinkHandle().LinkByName("veth0")
	if err != nil {
		t.Errorf("Failed to get link details: %s", err)
	}

	if h2Link.Attrs().MTU != 2000 {
		t.Errorf("Mismatching MTU")
	}

	if h2Link.Attrs().Group != 5678 {
		t.Errorf("Mismatching device group")
	}

	if !bytes.Equal(h2Link.Attrs().HardwareAddr, []byte{0, 0, 0, 0, 0, 2}) {
		t.Errorf("Mismatching MAC address")
	}
}
