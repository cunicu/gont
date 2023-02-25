// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"flag"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	co "github.com/stv0g/gont/pkg/options/capture"
	to "github.com/stv0g/gont/pkg/options/trace"
	"github.com/stv0g/gont/pkg/trace"
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

	n, err := g.NewNetwork(*nname,
		o.Customize[g.NetworkOption](globalNetworkOptions, t1)...)
	assert.NoError(t, err, "Failed to create network")

	err = t1.Start()
	assert.NoError(t, err, "Failed to start tracer")

	cmd, err := n.HostNode.RunGo("../test/tracee1")
	assert.NoError(t, err, "Failed to run sub-process")

	time.Sleep(100 * time.Millisecond)

	err = t1.Close()
	assert.NoError(t, err, "Failed to close tracer")

	myData := map[string]any{
		"Hello": "World",
	}

	filename := "test/tracee1/main.go"

	assert.NotNil(t, event, "No trace event received")
	assert.Equal(t, event.Type, "tracepoint", "Unexpected event type")
	assert.Equal(t, event.Message, "This is my first trace message: Hurra", "Wrong message")
	assert.Equal(t, event.Data, myData, "Mismatching data")
	assert.True(t, strings.HasSuffix(event.File, filename), "Mismatching filename")
	assert.NotEqual(t, event.Line, 0, "Empty line number")
	assert.Equal(t, event.PID, cmd.Process.Pid, "Wrong PID")
	assert.Equal(t, event.Function, "main.main", "Wrong function name")
}

func TestTraceInSameProcess(t *testing.T) {
	var event *trace.Event

	t1 := g.NewTracer(
		to.Callback(func(e trace.Event) {
			event = &e
		}),
	)

	err := t1.Start()
	assert.NoError(t, err, "Failed to start tracer")

	myData := map[string]string{
		"Hello": "World",
	}

	err = trace.PrintfWithData(myData, "This is my first trace message: %s", "Hurra")
	assert.NoError(t, err, "Failed to write trace")

	err = t1.Close()
	assert.NoError(t, err, "Failed to close tracer")

	filename := "pkg/trace_test.go"

	assert.NotNil(t, event, "No trace event received")
	assert.Equal(t, event.Type, "tracepoint", "Unexpected event type")
	assert.Equal(t, event.Message, "This is my first trace message: Hurra", "Wrong message")
	assert.Equal(t, event.Data, myData, "Mismatching data")
	assert.True(t, strings.HasSuffix(event.File, filename), "Mismatching filename")
	assert.NotEqual(t, event.Line, 0, "Empty line number")
	assert.Equal(t, event.PID, os.Getpid(), "Wrong PID")
	assert.Equal(t, event.Function, "github.com/stv0g/gont/pkg_test.TestTraceInSameProcess", "Wrong function name")
}

func TestTraceLog(t *testing.T) {
	var event *trace.Event

	t1 := g.NewTracer(
		to.Callback(func(e trace.Event) {
			event = &e
		}),
	)

	err := t1.Start()
	assert.NoError(t, err, "Failed to start tracer")

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
	assert.NoError(t, err, "Failed to close tracer")

	data := map[string]any{
		"string": "mystring",
		"number": int64(1234), // zap is adding zap.Int() as an int64 internally
	}

	_, filename, _, _ := runtime.Caller(0)

	function := "github.com/stv0g/gont/pkg_test.TestTraceLog"

	assert.NotNil(t, event, "No trace event received")
	assert.Equal(t, event.Type, "log", "Unexpected event type")
	assert.Equal(t, event.Data, data, "Mismatching data")
	assert.Equal(t, event.File, filename, "Mismatching filename")
	assert.NotEqual(t, event.Line, 0, "Empty line number")
	assert.Equal(t, event.PID, os.Getpid(), "Wrong PID")
	assert.Equal(t, event.Source, "my-test-logger", "Wrong logger name")
	assert.Equal(t, event.Level, uint8(zap.DebugLevel+2), "Wrong level")
	assert.Equal(t, event.Function, function, "Wrong function name")
}

func TestTraceWithCapture(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
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
	assert.NoError(t, err, "Failed to start")

	n, err := g.NewNetwork(*nname,
		o.Customize[g.NetworkOption](globalNetworkOptions, t1, c1)...)
	assert.NoError(t, err, "Failed to create network")

	h1, err := n.AddHost("h1")
	assert.NoError(t, err, "Failed to add host")

	for i := 0; i < 5; i++ {
		_, err = h1.RunGo("../test/tracee2", i)
		assert.NoError(t, err, "Failed to run tracee")
	}
}
