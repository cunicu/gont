// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"os"

	g "cunicu.li/gont/v2/pkg"
	"cunicu.li/gont/v2/pkg/trace"
)

// File writes tracing events in JSON format to the provided file handle.
// Each event is written to a new line.
type File struct {
	*os.File
}

func (f File) ApplyTracer(t *g.Tracer) {
	t.Files = append(t.Files, f.File)
}

func ToFile(f *os.File) File { return File{File: f} }

// Filename appends tracing events to a new or existing file with the provided filename.
type Filename string

func (fn Filename) ApplyTracer(c *g.Tracer) {
	c.Filenames = append(c.Filenames, string(fn))
}

func ToFilename(fn string) Filename { return Filename(fn) }

// Channel sends tracing events to the provided channel.
type Channel chan trace.Event

func (d Channel) ApplyTracer(t *g.Tracer) {
	t.Channels = append(t.Channels, d)
}

func ToChannel(ch chan trace.Event) Channel { return Channel(ch) }

// Callback provides a custom callback function which is called for each tracing event.
type Callback trace.EventCallback

func (cb Callback) ApplyTracer(t *g.Tracer) {
	t.Callbacks = append(t.Callbacks, trace.EventCallback(cb))
}

// Capture will write Tracepoints as fake packets to the capture.
//
// Checkout the included  Lua-based WireShark dissector to decode the fake
// tracing event packets.
//
// Start WireShark with the following options to load the dissector:
//
//	wireshark -Xlua_script:dissector.lua
//
// See ./dissector/dissector.lua
type Capture struct {
	*g.Capture
}

func (c Capture) ApplyTracer(t *g.Tracer) {
	t.Captures = append(t.Captures, c.Capture)
}

func ToCapture(c *g.Capture) Capture { return Capture{c} }
