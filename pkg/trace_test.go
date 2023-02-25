// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"flag"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

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
	var (
		event *trace.Event
		err   error
		n     *g.Network
	)

	t1 := g.NewTracer(
		to.Callback(func(e trace.Event) {
			event = &e
		}),
	)

	if n, err = g.NewNetwork(*nname,
		o.Customize[g.NetworkOption](globalNetworkOptions, t1)...,
	); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	if err := t1.Start(); err != nil {
		t.Fatalf("Failed to start tracer: %s", err)
	}

	cmd, err := n.HostNode.RunGo("../test/tracee1.go")
	if err != nil {
		t.Fatalf("Failed to run sub-process: %s", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := t1.Close(); err != nil {
		t.Fatalf("Failed to close tracer")
	}

	if event == nil {
		t.Fatal("No trace event received")
	}

	if event.Type != "tracepoint" {
		t.Fatalf("Unexpected event type: %s", event.Type)
	}

	if event.Message != "This is my first trace message: Hurra" {
		t.Fatal("Wrong message")
	}

	myData := map[string]any{
		"Hello": "World",
	}

	if !reflect.DeepEqual(event.Data, myData) {
		t.Fatalf("Mismatching data: %+#v != %+#v", event.Data, myData)
	}

	if !strings.HasSuffix(event.File, "test/tracee1.go") {
		t.Fatalf("Mismatching filename: %s != %s", event.File, "test/tracee1.go")
	}

	if event.Line == 0 {
		t.Fatal("Empty line number")
	}

	if event.PID != cmd.Process.Pid {
		t.Fatal("Wrong PID")
	}

	if event.Function != "main.main" {
		t.Fatalf("Wrong function name: %s != %s", event.Function, "main.main")
	}
}

func TestTracerInSameProcess(t *testing.T) {
	var event *trace.Event

	t1 := g.NewTracer(
		to.Callback(func(e trace.Event) {
			event = &e
		}),
	)

	if err := t1.Start(); err != nil {
		t.Fatalf("Failed to start tracer: %s", err)
	}

	myData := map[string]string{
		"Hello": "World",
	}

	if err := trace.PrintfWithData(myData, "This is my first trace message: %s", "Hurra"); err != nil {
		t.Fatalf("Failed to write trace: %s", err)
	}

	if err := t1.Close(); err != nil {
		t.Fatalf("Failed to close tracer")
	}

	if event == nil {
		t.Fatal("No trace event received")
	}

	if event.Type != "tracepoint" {
		t.Fatalf("Unexpected event type: %s", event.Type)
	}

	if !reflect.DeepEqual(event.Data, myData) {
		t.Fatal("Mismatching data")
	}

	_, filename, _, _ := runtime.Caller(0)
	if event.File != filename {
		t.Fatalf("Mismatching filename: %s != %s", event.File, filename)
	}

	if event.Line == 0 {
		t.Fatal("Empty line number")
	}

	if event.PID != os.Getpid() {
		t.Fatal("Wrong PID")
	}

	function := "github.com/stv0g/gont/pkg_test.TestTracerInSameProcess"
	if event.Function != function {
		t.Fatalf("Wrong function name: %s != %s", event.Function, function)
	}
}

func TestTracerLog(t *testing.T) {
	var event *trace.Event

	t1 := g.NewTracer(
		to.Callback(func(e trace.Event) {
			event = &e
		}),
	)

	if err := t1.Start(); err != nil {
		t.Fatalf("Failed to start tracer: %s", err)
	}

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

	if err := t1.Close(); err != nil {
		t.Fatalf("Failed to close tracer")
	}

	if event == nil {
		t.Fatal("No trace event received")
	}

	if event.Type != "log" {
		t.Fatalf("Unexpected event type: %s", event.Type)
	}

	data := map[string]any{
		"string": "mystring",
		"number": int64(1234), // zap is adding zap.Int() as an int64 internally
	}
	if !reflect.DeepEqual(event.Data, data) {
		t.Fatalf("Mismatching data: %+#v != %+#v", event.Data, data)
	}

	_, filename, _, _ := runtime.Caller(0)
	if event.File != filename {
		t.Fatalf("Mismatching filename: %s != %s", event.File, filename)
	}

	if event.Line == 0 {
		t.Fatal("Empty line number")
	}

	if event.PID != os.Getpid() {
		t.Fatal("Wrong PID")
	}

	if event.Source != "my-test-logger" {
		t.Fatal("Wrong logger name")
	}

	if event.Level != uint8(zap.DebugLevel+2) {
		t.Fatal("Wrong level")
	}

	function := "github.com/stv0g/gont/pkg_test.TestTracerLog"
	if event.Function != function {
		t.Fatalf("Wrong function name: %s != %s", event.Function, function)
	}
}

func TestTracerWithCapture(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		t.Skip("Requires WireShark")
	}

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

	if err := t1.Start(); err != nil {
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
		if _, err = h1.RunGo("../test/tracee2.go", i); err != nil {
			t.Fatalf("Failed to run tracee: %s", err)
		}
	}
}
