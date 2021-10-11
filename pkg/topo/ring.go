package topo

import (
	"net"

	g "github.com/stv0g/gont/pkg"
)

func Ring(n *g.Network, k, m int, gw net.IP, startAddr net.IP, mask net.IPMask, opts ...g.Option) ([]*g.Switch, []*g.Host, error) {
	// t := TopoOptions(opts...)

	sw, hs, err := Linear(n, 0, 0, opts)
	if err != nil {
		return nil, nil, err
	}

	// Connect first and last switch
	// if err := t.LinkHosts(g, k, sw[0], sw[k-1], opts...); err != nil {
	// 	return nil, nil, err
	// }

	return sw, hs, nil
}
