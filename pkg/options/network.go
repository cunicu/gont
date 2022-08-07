package options

import (
	g "github.com/stv0g/gont/pkg"
)

type NSPrefix string
type Persistent bool
type CaptureNetwork struct {
	*g.Capture
}

func (pfx NSPrefix) Apply(n *g.Network) {
	n.NSPrefix = string(pfx)
}

func (p Persistent) Apply(n *g.Network) {
	n.Persistent = bool(p)
}

func (c CaptureNetwork) Apply(n *g.Network) {
	n.Captures = append(n.Captures, c.Capture)
}

func CaptureAll(opts ...g.Option) CaptureNetwork {
	return CaptureNetwork{Capture(opts...)}
}
