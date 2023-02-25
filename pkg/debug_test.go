// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	do "github.com/stv0g/gont/pkg/options/debug"
	to "github.com/stv0g/gont/pkg/options/trace"
	"github.com/stv0g/gont/pkg/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

func TestDebugger(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		t.Skip("Requires WireShark")
	}

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
			err := enc.Encode(e)
			assert.NoError(t, err, "Failed to encode")
		}),
	)

	err := s.Start()
	assert.NoError(t, err, "Failed to start tracer")

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
		// 	do.File("/home/stv0g/workspace/gont/test/tracee2/main.go"), do.Line(43),
		// 	do.Data("my user data"),
		// ),
		do.ToTracer(s),
	)

	n, err := g.NewNetwork(*nname,
		o.Customize[g.NetworkOption](globalNetworkOptions,
			o.RedirectToLog(true),
			d,
		)...)
	assert.NoError(t, err, "Failed to create network")

	h1, err := n.AddHost("h1")
	assert.NoError(t, err, "Failed to add host")

	_, err = h1.StartGo("../test/tracee2", 1)
	assert.NoError(t, err, "Failed to run tracee")

	_, err = h1.StartGo("../test/tracee2", 2)
	assert.NoError(t, err, "Failed to run tracee")

	err = d.WriteVSCodeConfigs("", false)
	assert.NoError(t, err, "Failed to write VSCode config")

	select {}
}
