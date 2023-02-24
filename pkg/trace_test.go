// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"flag"
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	co "github.com/stv0g/gont/pkg/options/capture"
	to "github.com/stv0g/gont/pkg/options/trace"
	"github.com/stv0g/gont/pkg/trace"
)

//nolint:gochecknoglobals
var captureSocketAddr = flag.String("trace-socket", "tcp:[::1]:42125", "Listen address for capture socket")

func TestTracer(t *testing.T) {
	var (
		err error
		n   *g.Network
		h1  *g.Host
	)

	c1 := g.NewCapture(
		co.ListenAddr(*captureSocketAddr),
	)

	t1 := g.NewTracer(
		to.ToFilename("trace.log"),
		to.ToCapture(c1),
	)

	if err := t1.StartLocal(); err != nil {
		t.Fatalf("Failed to start: %s", err)
	}

	if n, err = g.NewNetwork(*nname,
		o.Customize[g.NetworkOption](globalNetworkOptions, t1, c1)...,
	); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	if h1, err = n.AddHost("h1",
		o.RedirectToLog(true),
	); err != nil {
		t.Fatalf("Failed to add host: %s", err)
	}

	for i := 0; i < 5; i++ {
		if _, err = h1.Run("../test/tracee/tracee", i); err != nil {
			t.Fatalf("Failed to run tracee: %s", err)
		}

		if err := trace.PrintfWithData(i, "Hello from test process"); err != nil {
			t.Fatalf("Failed to print: %s", err)
		}
	}
}
