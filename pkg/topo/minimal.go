package topo

import (
	g "github.com/stv0g/gont/pkg"
)

func Minimal(n *g.Network, opts ...g.Option) (*g.Switch, []*g.Host, error) {
	return SingleSwitch(n, 2, opts)
}
