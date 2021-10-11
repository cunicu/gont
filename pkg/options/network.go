package options

import (
	g "github.com/stv0g/gont/pkg"
)

type NSPrefix string
type Persistent bool

func (pfx NSPrefix) Apply(n *g.Network) {
	n.NSPrefix = string(pfx)
}

func (p Persistent) Apply(n *g.Network) {
	n.Persistent = bool(p)
}

func DefaultNetwork() (*g.Network, error) {
	return g.NewNetwork("",
		MTU(1500))
}
