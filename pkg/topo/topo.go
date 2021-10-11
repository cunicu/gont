package topo

import (
	"net"

	g "github.com/stv0g/gont/pkg"
)

type Topo struct {
	Dimensions Vector

	CreateHost CreateHost
	LinkHosts  LinkHosts

	StartAddress net.IP
}

type Vector []int
type CreateHost func(n *g.Network, coord Vector, opts ...g.Option) (*g.Host, error)
type LinkHosts func(n *g.Network, a, b *g.Host, opts ...g.Option) error

func TopoOptions(opts ...g.Option) *Topo {
	topo := &Topo{}

	for _, opt := range opts {
		if topt, ok := opt.(TopoOption); ok {
			topt.apply(topo)
		}
	}

	return topo
}
