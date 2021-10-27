package topo

import (
	"fmt"
	"net"

	g "github.com/stv0g/gont/pkg"
)

func SingleSwitch(n *g.Network, m int, opts ...g.Option) (*g.Switch, []*g.Host, error) {
	sw, err := n.AddSwitch("sw", opts...)
	if err != nil {
		return nil, nil, err
	}

	t := TopoOptions(opts...)

	hs := []*g.Host{}

	for i := 0; i < m; i++ {
		a := t.StartAddress
		a[4] += byte(i)

		opts = append(opts, g.Interface{
			Port: g.Port{
				Name: "veth0",
				Node: sw,
			},
			Addresses: []net.IPNet{}, // TODO: assign addresses here
		})

		hName := fmt.Sprintf("h%d", i)
		h, err := n.AddHost(hName, opts...)
		if err != nil {
			return nil, nil, err
		}

		hs = append(hs, h)
	}

	return sw, hs, nil
}
