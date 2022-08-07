package options

import (
	g "github.com/stv0g/gont/pkg"
)

type NSPrefix string
type Persistent bool
type captureNetwork struct {
	*g.Capture
}

func (pfx NSPrefix) Apply(n *g.Network) {
	n.NSPrefix = string(pfx)
}

func (p Persistent) Apply(n *g.Network) {
	n.Persistent = bool(p)
}

func (c captureNetwork) Apply(n *g.Network) {
	n.Captures = append(n.Captures, c.Capture)
}

func CaptureNetwork(opts ...g.Option) captureNetwork {
	return captureNetwork{Capture(opts...)}
}
