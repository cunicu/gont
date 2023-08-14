// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Limit debugger support to platforms which are supported by Delve
// See: https://github.com/go-delve/delve/blob/master/pkg/proc/native/support_sentinel_linux.go
//go:build linux && (amd64 || arm64 || 386)

package debug

import (
	g "cunicu.li/gont/v2/pkg"
)

// ListenAddr opens a TCP socket listening for DAP connections.
//
// This allows attaching a DAP-compatible IDE like VScode to
// the debugger.
//
// See: https://microsoft.github.io/debug-adapter-protocol/
type ListenAddr string

func (l ListenAddr) ApplyDebugger(d *g.Debugger) {
	d.ListenAddr = string(l)
}

// BreakOnEntry will stop the debugged process directly after
// launch. So that the user can set break- and watchpoints before
// continuing the execution.
type BreakOnEntry bool

func (b BreakOnEntry) ApplyDebugger(d *g.Debugger) {
	d.BreakOnEntry = bool(b)
}

// InfoDirectory appends a debug info directory which is
// used by the debugger to parse debug information.
//
// See also: https://sourceware.org/gdb/onlinedocs/gdb/Separate-Debug-Files.html
type InfoDirectory string

func (b InfoDirectory) ApplyDebugger(d *g.Debugger) {
	d.DebugInfoDirectories = append(d.DebugInfoDirectories, string(b))
}

// Tracer attaches a tracer to the debugger.
//
// All tracepoints will be send as trace events to the tracer.
// This allows for recording tracepoints to files, or handling them
// via a channel or callback.
//
// It also enables the recording of debugger tracepoints to a Capture
// and hence recording them to a PCAPng file. This enables the user
// to analyze both network data, tracing events and debugger break/watchpoints
// in a single stream of events in WireShark.
type Tracer struct {
	*g.Tracer
}

func (t Tracer) ApplyDebugger(d *g.Debugger) {
	d.Tracers = append(d.Tracers, t.Tracer)
}

func ToTracer(t *g.Tracer) Tracer { return Tracer{t} }
