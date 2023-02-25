// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"encoding/json"
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	do "github.com/stv0g/gont/pkg/options/debug"
	to "github.com/stv0g/gont/pkg/options/trace"
	"github.com/stv0g/gont/pkg/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

func TestDebugger(t *testing.T) {
	t.Skip("Requires interactive debugger")

	var (
		err error
		n   *g.Network
		h1  *g.Host
	)

	// c := g.NewCapture(
	// 	co.Listener("tcp:0.0.0.0:5678"),
	// )

	s := g.NewTracer(
		// to.ToCapture(c),
		to.Callback(func(e trace.Event) {
			enc := json.NewEncoder(&zapio.Writer{
				Log:   zap.L(),
				Level: zap.InfoLevel,
			})
			enc.SetIndent("", "  ")
			if err := enc.Encode(e); err != nil {
				t.Fatal(err)
			}
		}),
	)

	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	d := g.NewDebugger(
		do.ListenAddr(":1234"),
		do.BreakOnEntry(true),
		// g.NewTracepoint(
		// 	do.FunctionsRegex("main.myTime"),
		// 	do.Message("Variable i is {i}"),
		// 	do.Condition("i >= 4"),
		// 	do.Stacktrace(1),
		// 	do.LoadArguments(do.LoadConfig(
		// 		do.FollowPointers(true),
		// 		do.MaxStructFields(10),
		// 		do.MaxStringLen(10),
		// 	)),
		// 	do.LoadLocals(do.LoadConfig(
		// 		do.FollowPointers(true),
		// 		do.MaxStructFields(10),
		// 		do.MaxStringLen(10),
		// 	)),
		// ),
		// g.NewTracepoint(
		// 	do.Location("/main\\..*/"),
		// ),
		// g.NewTracepoint(
		// 	do.Name("My watchpoint"), do.Description("hallo", "welt"),
		// 	do.Watch("i", api.WatchWrite),
		// 	do.File("/home/stv0g/workspace/gont/test/tracee2.go"), do.Line(43),
		// 	do.Data("my user data"),
		// ),
		do.ToTracer(s),
	)

	if n, err = g.NewNetwork(*nname,
		// o.Customize[g.NetworkOption](globalNetworkOptions,
		o.RedirectToLog(true),
		d,
		// )...,
	); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	if h1, err = n.AddHost("h1"); err != nil {
		t.Fatalf("Failed to add host: %s", err)
	}

	if _, err = h1.StartGo("../test/tracee2.go", 1); err != nil {
		t.Fatalf("Failed to run tracee: %s", err)
	}

	if _, err = h1.StartGo("../test/tracee2.go", 2); err != nil {
		t.Fatalf("Failed to run tracee: %s", err)
	}

	if err := d.WriteVSCodeConfigs("", false); err != nil {
		t.Fatal(err)
	}

	select {}
}
