package options

import (
	g "github.com/stv0g/gont/pkg"
)

type Persistent bool

func (p Persistent) Apply(n *g.Network) {
	n.Persistent = bool(p)
}

func DefaultNetwork() (*g.Network, error) {
	return g.NewNetwork("",
		MTU(1500))
}
