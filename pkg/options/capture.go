package options

import (
	"os"

	g "github.com/stv0g/gont/pkg"
)

type CaptureLength int

func (sl CaptureLength) Apply(c *g.Capture) {
	c.CaptureLength = int(sl)
}

type Promisc bool

func (p Promisc) Apply(c *g.Capture) {
	c.Promisc = bool(p)
}

type FilterInterfaces g.CaptureFilterInterfaceFunc

func (f FilterInterfaces) Apply(c *g.Capture) {
	c.Filter = g.CaptureFilterInterfaceFunc(f)
}

type BPFilter string

func (bpf BPFilter) Apply(c *g.Capture) {
	c.BPFilter = string(bpf)
}

type Comment string

func (d Comment) Apply(c *g.Capture) {
	c.Comment = string(d)
}

type File struct {
	*os.File
}

func (f File) Apply(c *g.Capture) {
	c.File = f.File
}

type Filename string

func (fn Filename) Apply(c *g.Capture) {
	c.Filename = string(fn)
}

func Capture(opts ...g.Option) *g.Capture {
	c := g.NewCapture()

	for _, o := range opts {
		if o, ok := o.(g.CaptureOption); ok {
			o.Apply(c)
		}
	}

	return c
}
