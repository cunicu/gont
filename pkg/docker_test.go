// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"bytes"
	"net/url"
	"strings"
	"testing"

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	co "cunicu.li/gont/v2/pkg/options/cmd"
	"github.com/stretchr/testify/require"
)

func TestDocker(t *testing.T) {
	t.Skip("Test is currently broken")

	n, err := g.NewNetwork(*nname)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	sw, err := n.AddSwitch("sw")
	require.NoError(t, err, "Failed to create switch")

	// h1 is a normal Gont node
	h1, err := n.AddHost("h1",
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.0.1/24"),
			o.AddressIP("fc::1/64")))
	require.NoError(t, err, "Failed to create host")

	// h2 is a Docker container
	outp := &bytes.Buffer{}
	_, err = n.HostNode.Run("docker", "run", "--detach", "nginx",
		co.Stdout(outp),
	)
	require.NoError(t, err, "Failed to start Docker container")

	id := strings.TrimSpace(outp.String())

	t.Logf("Started nginx Docker container with id %s", id)

	h2, err := n.AddHost("h2",
		o.ExistingDockerContainer(id),
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.0.2/24"),
			o.AddressIP("fc::2/64")),
	)
	require.NoError(t, err, "Failed to create host")

	_, err = h2.Run("hostname")
	require.NoError(t, err, "Failed to run")

	u, err := url.Parse("http://h2/")
	require.NoError(t, err, "Failed to parse URL")

	_, err = h1.Run("curl", u)
	require.NoError(t, err, "Failed to run")
	_, err = h2.Ping(h1)
	require.NoError(t, err, "Failed to run")
}
