// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"testing"

	g "cunicu.li/gont/v2/pkg"
	"github.com/stretchr/testify/require"
)

func TestLink(t *testing.T) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h1, err := n.AddHost("h1")
	require.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2")
	require.NoError(t, err, "Failed to create host")

	err = n.AddLink(
		g.NewInterface("veth0", h1),
		g.NewInterface("veth0", h2))
	require.NoError(t, err, "Failed to link nodes")
}
