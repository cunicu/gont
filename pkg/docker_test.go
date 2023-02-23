// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"bytes"
	"net/url"
	"strings"
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	co "github.com/stv0g/gont/pkg/options/cmd"
)

func TestDocker(t *testing.T) {
	// Test is broken
	t.Skip()

	var (
		err    error
		n      *g.Network
		sw     *g.Switch
		h1, h2 *g.Host
	)

	if n, err = g.NewNetwork(*nname, globalNetworkOptions...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
	defer n.Close()

	if sw, err = n.AddSwitch("sw"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	// h1 is a normal Gont node
	if h1, err = n.AddHost("h1",
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.0.1/24"),
			o.AddressIP("fc::1/64")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	// h2 is a Docker container
	outp := &bytes.Buffer{}
	_, err = n.HostNode.Run("docker", "run", "--detach", "nginx",
		co.Stdout(outp),
	)
	if err != nil {
		t.Fatalf("Failed to start Docker container")
	}

	id := strings.TrimSpace(outp.String())

	t.Logf("Started nginx Docker container with id %s", id)

	if h2, err = n.AddHost("h2",
		o.ExistingDockerContainer(id),
		g.NewInterface("veth0", sw,
			o.AddressIP("10.0.0.2/24"),
			o.AddressIP("fc::2/64")),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if _, err := h2.Run("hostname"); err != nil {
		t.Fatalf("Failed to run: %s", err)
	}

	u, err := url.Parse("http://h2/")
	if err != nil {
		t.Fail()
	}

	if _, err := h1.Run("curl", u); err != nil {
		t.Fatalf("Failed to run: %s", err)
	}
	if _, err := h2.Ping(h1); err != nil {
		t.Fatalf("Failed to run: %s", err)
	}
}
