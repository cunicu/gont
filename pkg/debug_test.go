// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Limit debugger support to platforms which are supported by Delve
// See: https://github.com/go-delve/delve/blob/master/pkg/proc/native/support_sentinel_linux.go
//go:build linux && (amd64 || arm64 || 386)

package gont_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	do "cunicu.li/gont/v2/pkg/options/debug"
	to "cunicu.li/gont/v2/pkg/options/trace"
	"cunicu.li/gont/v2/pkg/trace"
	"github.com/go-delve/delve/service/api"
	"github.com/go-delve/delve/service/dap/daptest"
	"github.com/stretchr/testify/require"
)

func TestDebugBreakpointHostNode(t *testing.T) {
	tps := []trace.Event{}

	v := g.NewTracer(
		to.Callback(func(tp trace.Event) { tps = append(tps, tp) }),
	)

	d := g.NewDebugger(
		do.ToTracer(v),
		g.NewTracepoint(
			do.File("../test/debugee1/main.go"),
			do.Line(22),
			do.Variable("s")))

	n, err := g.NewNetwork("", d)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	_, err = n.HostNode.RunGo("../test/debugee1", o.BuildFlagsDebug)
	require.NoError(t, err, "Failed to run")

	err = v.Close()
	require.NoError(t, err, "Failed to close tracer")

	require.Len(t, tps, 1)
	require.Equal(t, "Hello World", tps[0].Breakpoint.Variable("s"))
}

func TestDebugBreakpoint(t *testing.T) {
	tps := []trace.Event{}

	v := g.NewTracer(
		to.Callback(func(tp trace.Event) { tps = append(tps, tp) }),
	)

	d := g.NewDebugger(
		do.ToTracer(v),
		g.NewTracepoint(
			do.File("../test/debugee1/main.go"),
			do.Line(22),
			do.Data(1337),
			do.Stacktrace(5),
			do.Message("Variable t.A has the value {t.A}"),
			do.LoadLocals(
				do.MaxStringLen(100),
				do.MaxVariableRecurse(5),
				do.MaxArrayValues(100),
				do.MaxStructFields(100),
				do.FollowPointers(true))))

	n, err := g.NewNetwork("", d)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h1, err := n.AddHost("h1")
	require.NoError(t, err, "Failed to add host")

	_, err = h1.RunGo("../test/debugee1",
		o.BuildFlagsDebug)
	require.NoError(t, err, "Failed to run")

	err = v.Close()
	require.NoError(t, err, "Failed to close tracer")

	require.Len(t, tps, 1)

	tp := tps[0]
	require.True(t, strings.HasSuffix(tp.File, "test/debugee1/main.go"))
	require.Equal(t, 22, tp.Line)
	require.Equal(t, "Variable t.A has the value 555", tp.Message)
	require.Equal(t, "breakpoint", tp.Type)
	require.Equal(t, 1337, tp.Data)

	require.NotNil(t, tp.Breakpoint)
	require.Equal(t, 1, tp.Breakpoint.ID)
	require.EqualValues(t, 1, tp.Breakpoint.TotalHitCount)
	require.Equal(t, "Hello World", tp.Breakpoint.Local("s"))
}

func TestDebugBreakpointLocation(t *testing.T) {
	tes := []trace.Event{}

	v := g.NewTracer(
		to.Callback(func(te trace.Event) { tes = append(tes, te) }),
	)

	d := g.NewDebugger(
		do.ToTracer(v),
		g.NewTracepoint(
			do.FunctionNameRegex("main.myFunction"),
			do.Data(2000),
			do.Stacktrace(5),
			do.LoadArguments(
				do.MaxStringLen(100),
				do.MaxVariableRecurse(5),
				do.MaxArrayValues(100),
				do.MaxStructFields(100),
				do.FollowPointers(true))))

	n, err := g.NewNetwork("", d)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h1, err := n.AddHost("h1")
	require.NoError(t, err, "Failed to add host")

	_, err = h1.RunGo("../test/debugee1",
		o.BuildFlagsDebug)
	require.NoError(t, err, "Failed to run")

	err = v.Close()
	require.NoError(t, err, "Failed to close tracer")

	assert := func(expectedHitCount uint64, tp trace.Event) {
		require.True(t, strings.HasSuffix(tp.File, "test/debugee1/main.go"))
		require.Equal(t, 27, tp.Line)
		require.Equal(t, "Hit breakpoint 1: /main.myFunction/ (0)", tp.Message)
		require.Equal(t, "breakpoint", tp.Type)
		require.Equal(t, 2000, tp.Data)

		bp := tp.Breakpoint
		require.NotNil(t, bp)
		require.Equal(t, 1, bp.ID)
		require.EqualValues(t, expectedHitCount, bp.TotalHitCount)
		require.Equal(t, "Hello World", bp.Argument("s"))
	}

	require.Len(t, tes, 2)
	assert(1, tes[0])
	assert(2, tes[1])
}

func TestDebugWatchpoint(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		t.Skip("Test broken on Azure")
	}

	tps := []trace.Event{}

	v := g.NewTracer(
		to.Callback(func(tp trace.Event) { tps = append(tps, tp) }),
	)

	d := g.NewDebugger(
		do.ToTracer(v),
		g.NewTracepoint(
			do.File("../test/debugee1/main.go"),
			do.Line(30),
			do.Data(1337),
			do.Message("i has the value '{i}'"),
			do.Watch("i", api.WatchWrite),
			do.Variable("t")))

	n, err := g.NewNetwork("", d)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h1, err := n.AddHost("h1")
	require.NoError(t, err, "Failed to add host")

	_, err = h1.RunGo("../test/debugee1",
		o.BuildFlagsDebug)
	require.NoError(t, err, "Failed to run")

	err = v.Close()
	require.NoError(t, err, "Failed to close tracer")

	assert := func(expectedHitCount uint64, expectedValue string, tp trace.Event) {
		require.True(t, strings.HasSuffix(tp.File, "test/debugee1/main.go"))
		require.Equal(t, 30, tp.Line)
		require.Equal(t, fmt.Sprintf("i has the value '%s'", expectedValue), tp.Message)
		require.Equal(t, "watchpoint", tp.Type)
		require.Equal(t, 1337, tp.Data)

		bp := tp.Breakpoint
		require.NotNil(t, bp)
		require.EqualValues(t, expectedHitCount, bp.TotalHitCount)
		require.Equal(t, expectedValue, bp.Variable("i"))
	}

	require.Len(t, tps, 2)
	assert(1, "1337", tps[0])
	assert(2, "2674", tps[1])
}

func TestDebugListener(t *testing.T) {
	tps := []trace.Event{}

	listenAddr := "[::1]:1234"

	v := g.NewTracer(
		to.Callback(func(tp trace.Event) { tps = append(tps, tp) }),
	)

	d := g.NewDebugger(
		do.ToTracer(v),
		do.ListenAddr(listenAddr),
		do.BreakOnEntry(true),
		g.NewTracepoint(
			do.File("../test/debugee1/main.go"),
			do.Line(21),
			do.Variable("s")))

	n, err := g.NewNetwork("", d)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	c, err := n.HostNode.StartGo("../test/debugee1", o.BuildFlagsDebug)
	require.NoError(t, err, "Failed to run")

	testDAP(t, listenAddr)

	err = c.Wait()
	require.NoError(t, err, "Failed to wait")

	err = v.Close()
	require.NoError(t, err, "Failed to close tracer")

	require.Len(t, tps, 0)
}

func testDAP(t *testing.T, listenAddr string) {
	client := daptest.NewClient(listenAddr)
	defer client.Close()

	// Get list of threads
	client.ThreadsRequest()
	threadResponse := client.ExpectThreadsResponse(t)

	require.True(t, threadResponse.Success)
	require.True(t, len(threadResponse.Body.Threads) >= 1)

	currentThread := threadResponse.Body.Threads[0]

	// Continue
	client.ContinueRequest(currentThread.Id)
	client.ExpectContinueResponse(t)

	// Check halt
	client.ExpectStoppedEvent(t)

	// Retrieve current frame
	client.StackTraceRequest(currentThread.Id, 0, 1)
	stResponse := client.ExpectStackTraceResponse(t)

	require.Equal(t, "runtime.main", stResponse.Body.StackFrames[0].Name)

	// Continue
	client.ContinueRequest(currentThread.Id)
	client.ExpectContinueResponse(t)

	// Check halt
	client.ExpectStoppedEvent(t)

	// Retrieve current frame
	client.StackTraceRequest(currentThread.Id, 0, 1)
	stResponse = client.ExpectStackTraceResponse(t)

	require.Equal(t, "main.main", stResponse.Body.StackFrames[0].Name)
	require.Equal(t, "main.go", stResponse.Body.StackFrames[0].Source.Name)
	require.Equal(t, 21, stResponse.Body.StackFrames[0].Line)

	// Set a new breakpoint
	file, err := filepath.Abs("../test/debugee1/main.go")
	require.NoError(t, err)

	client.SetBreakpointsRequest(file, []int{22})
	bpResponse := client.ExpectSetBreakpointsResponse(t)

	require.True(t, bpResponse.Success)
	require.Len(t, bpResponse.Body.Breakpoints, 1)

	bp := bpResponse.Body.Breakpoints[0]
	require.True(t, bp.Verified, bp.Message)

	// Continue
	client.ContinueRequest(currentThread.Id)
	client.ExpectContinueResponse(t)

	// Check halt
	client.ExpectStoppedEvent(t)

	// Retrieve current frame
	client.StackTraceRequest(currentThread.Id, 0, 1)
	stResponse = client.ExpectStackTraceResponse(t)

	require.Equal(t, "main.main", stResponse.Body.StackFrames[0].Name)
	require.Equal(t, "main.go", stResponse.Body.StackFrames[0].Source.Name)
	require.Equal(t, 22, stResponse.Body.StackFrames[0].Line)
}
