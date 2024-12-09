// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	co "cunicu.li/gont/v2/pkg/options/capture"
	to "cunicu.li/gont/v2/pkg/options/trace"
	"cunicu.li/gont/v2/pkg/trace"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals
var captureSocketAddr = flag.String("trace-socket", "tcp:[::1]:42125", "Listen address for capture socket")

func TestTraceSubProcess(t *testing.T) {
	var event *trace.Event

	t1 := g.NewTracer(
		to.Callback(func(e trace.Event) {
			event = &e
		}),
	)

	n, err := g.NewNetwork(*nname, t1)
	require.NoError(t, err, "Failed to create network")

	hn, err := n.AddHost("host", o.HostNamespace)
	require.NoError(t, err, "Failed to create host node")

	err = t1.Start()
	require.NoError(t, err, "Failed to start tracer")

	cmd, err := hn.RunGo("../test/tracee1")
	require.NoError(t, err, "Failed to run sub-process")

	time.Sleep(100 * time.Millisecond)

	err = t1.Close()
	require.NoError(t, err, "Failed to close tracer")

	myData := map[string]any{
		"Hello": "World",
	}

	filename := "test/tracee1/main.go"

	require.NotNil(t, event, "No trace event received")
	require.Equal(t, event.Type, "tracepoint", "Unexpected event type")
	require.Equal(t, event.Message, "This is my first trace message: Hurra", "Wrong message")
	require.Equal(t, event.Data, myData, "Mismatching data")
	require.True(t, strings.HasSuffix(event.File, filename), "Mismatching filename")
	require.NotZero(t, event.Line, "Empty line number")
	require.Equal(t, event.PID, cmd.Process.Pid, "Wrong PID")
	require.Equal(t, event.Function, "main.main", "Wrong function name")
}

func TestTraceInSameProcess(t *testing.T) {
	var event *trace.Event

	t1 := g.NewTracer(
		to.Callback(func(e trace.Event) {
			event = &e
		}),
	)

	err := t1.Start()
	require.NoError(t, err, "Failed to start tracer")

	myData := map[string]string{
		"Hello": "World",
	}

	err = trace.PrintfWithData(myData, "This is my first trace message: %s", "Hurra")
	require.NoError(t, err, "Failed to write trace")

	err = t1.Close()
	require.NoError(t, err, "Failed to close tracer")

	filename := "pkg/trace_test.go"

	require.NotNil(t, event, "No trace event received")
	require.Equal(t, event.Type, "tracepoint", "Unexpected event type")
	require.Equal(t, event.Message, "This is my first trace message: Hurra", "Wrong message")
	require.Equal(t, event.Data, myData, "Mismatching data")
	require.True(t, strings.HasSuffix(event.File, filename), "Mismatching filename")
	require.NotZero(t, event.Line, "Empty line number")
	require.Equal(t, event.PID, os.Getpid(), "Wrong PID")
	require.Equal(t, event.Function, "cunicu.li/gont/v2/pkg_test.TestTraceInSameProcess", "Wrong function name")
}

func TestTraceLog(t *testing.T) {
	var event *trace.Event

	t1 := g.NewTracer(
		to.Callback(func(e trace.Event) {
			event = &e
		}),
	)

	err := t1.Start()
	require.NoError(t, err, "Failed to start tracer")

	logger := zap.L()

	// Add the tracing option which emits a trace event for each log message
	logger = logger.WithOptions(trace.Log())

	// Add the caller info which gets also included in the trace event
	logger = logger.WithOptions(zap.AddCaller())

	// Give the logger some name which is added as the Source field to the trace event
	logger = logger.Named("my-test-logger")

	logger.Debug("This is a log message",
		zap.String("string", "mystring"),
		zap.Int("number", 1234))

	err = t1.Close()
	require.NoError(t, err, "Failed to close tracer")

	data := map[string]any{
		"string": "mystring",
		"number": int64(1234), // zap is adding zap.Int() as an int64 internally
	}

	_, filename, _, _ := runtime.Caller(0)

	function := "cunicu.li/gont/v2/pkg_test.TestTraceLog"

	require.NotNil(t, event, "No trace event received")
	require.Equal(t, event.Type, "log", "Unexpected event type")
	require.Equal(t, event.Data, data, "Mismatching data")
	require.Equal(t, event.File, filename, "Mismatching filename")
	require.NotZero(t, event.Line, "Empty line number")
	require.Equal(t, event.PID, os.Getpid(), "Wrong PID")
	require.Equal(t, event.Source, "my-test-logger", "Wrong logger name")
	require.Equal(t, event.Level, uint8(zap.DebugLevel+2), "Wrong level")
	require.Equal(t, event.Function, function, "Wrong function name")
}

func TestTraceWithCapture(t *testing.T) {
	if _, ok := os.LookupEnv("WITH_WIRESHARK"); !ok {
		t.Skip("Requires WireShark")
	}

	c1 := g.NewCapture(
		co.ListenAddr(*captureSocketAddr),
	)

	t1 := g.NewTracer(
		to.ToFilename("trace.log"),
		to.ToCapture(c1),
	)

	err := t1.Start()
	require.NoError(t, err, "Failed to start")

	n, err := g.NewNetwork(*nname, t1, c1)
	require.NoError(t, err, "Failed to create network")

	h1, err := n.AddHost("h1")
	require.NoError(t, err, "Failed to add host")

	for i := 0; i < 5; i++ {
		_, err = h1.RunGo("../test/tracee2", i)
		require.NoError(t, err, "Failed to run tracee")
	}
}

func TestTraceDissector(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		t.Skip("Test flaky on Azure")
	}

	tmpPCAP, err := os.CreateTemp(t.TempDir(), "gont-capture-*.pcapng")
	require.NoError(t, err, "Failed to open temporary file")

	c1 := g.NewCapture(
		co.ToFile(tmpPCAP),
	)

	t1 := g.NewTracer(
		to.ToCapture(c1),
	)

	err = t1.Start()
	require.NoError(t, err, "Failed to start tracer")

	err = trace.PrintfWithData(1237, "This is my first trace message: %s", "Hurra")
	require.NoError(t, err, "Failed to write trace")

	err = t1.Close()
	require.NoError(t, err, "Failed to close tracer")

	err = c1.Close()
	require.NoError(t, err, "Failed to close capture")

	t.Logf("PCAPng file: %s", tmpPCAP.Name())

	err = tmpPCAP.Close()
	require.NoError(t, err, "Failed to close")

	c := exec.Command("tshark", "-Xlua_script:../dissector/dissector.lua", "-r", tmpPCAP.Name(), "-T", "json") //nolint:gosec

	buf := &bytes.Buffer{}
	c.Stdout = buf

	err = c.Run()
	require.NoError(t, err, "Failed to run tshark")

	// TODO: Figure out why its not flushed
	time.Sleep(300 * time.Millisecond)

	var tsharkOutput []TsharkOutput
	err = json.Unmarshal(buf.Bytes(), &tsharkOutput)
	require.NoError(t, err, "Failed to parse Tshark JSON output")

	require.Len(t, tsharkOutput, 1)

	trace := tsharkOutput[0].Source.Layers.Trace
	require.Equal(t, trace.Message, "This is my first trace message: Hurra")
	require.Equal(t, trace.Type, "tracepoint")
	require.Equal(t, trace.Function, "cunicu.li/gont/v2/pkg_test.TestTraceDissector")
	require.Equal(t, trace.Data, fmt.Sprint(1237))
	require.Equal(t, trace.Pid, fmt.Sprint(os.Getpid()))
	require.True(t, strings.HasSuffix(trace.File, "gont/pkg/trace_test.go"))
}
