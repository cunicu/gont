package trace

import (
	"os"

	g "github.com/stv0g/gont/pkg"
	"github.com/stv0g/gont/pkg/trace"
)

// File writes all captured packets to a file handle
type File struct {
	*os.File
}

func (f File) Apply(t *g.Tracer) {
	t.Files = append(t.Files, f.File)
}

func ToFile(f *os.File) File { return File{File: f} }

// Filename writes all captured packets to a PCAPng file
type Filename string

func (fn Filename) Apply(c *g.Tracer) {
	c.Filenames = append(c.Filenames, string(fn))
}

func ToFilename(fn string) Filename { return Filename(fn) }

// Channel sends all captured packets to the provided channel.
type Channel chan trace.Event

func (d Channel) Apply(t *g.Tracer) {
	t.Channels = append(t.Channels, d)
}

func ToChannel(ch chan trace.Event) Channel { return Channel(ch) }

// Callback provides a custom callback function which is called for each captured packet
type Callback trace.EventCallback

func (cb Callback) Apply(t *g.Tracer) {
	t.Callbacks = append(t.Callbacks, trace.EventCallback(cb))
}

// Capture will write Tracepoints as fake packets to the capture
type Capture struct {
	*g.Capture
}

func (c Capture) Apply(t *g.Tracer) {
	t.Captures = append(t.Captures, c.Capture)
}

func ToCapture(c *g.Capture) Capture { return Capture{c} }
