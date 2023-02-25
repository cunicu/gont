// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

// TestPing performs and end-to-end ping test
// between two hosts on a switched topology
//
//	h1 <-> sw1 <-> sw2 <-> h2
func TestPingCascadedSwitches(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.Close()

	sw1, err := n.AddSwitch("sw1")
	require.NoError(t, err, "Failed to add switch")

	sw2, err := n.AddSwitch("sw2")
	require.NoError(t, err, "Failed to add switch")

	h1, err := n.AddHost("h1",
		g.NewInterface("veth0", sw1,
			o.AddressIP("10.0.0.1/24")))
	require.NoError(t, err, "Failed to add host")

	h2, err := n.AddHost("h2",
		g.NewInterface("veth0", sw2,
			o.AddressIP("10.0.0.2/24")))
	require.NoError(t, err, "Failed to add host")

	err = n.AddLink(
		g.NewInterface("br-sw2", sw1),
		g.NewInterface("br-sw1", sw2))
	require.NoError(t, err, "Failed to add link")

	err = g.TestConnectivity(h1, h2)
	require.NoError(t, err, "Failed to check connectivity")
}
