// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	g "github.com/stv0g/gont/v2/pkg"
)

func TestLink(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.Close()

	h1, err := n.AddHost("h1")
	require.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2")
	require.NoError(t, err, "Failed to create host")

	err = n.AddLink(
		g.NewInterface("veth0", h1),
		g.NewInterface("veth0", h2))
	require.NoError(t, err, "Failed to link nodes")
}
