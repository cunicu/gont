package topo

import (
	"fmt"
	"net"

	g "github.com/stv0g/gont/pkg"
)

func Linear(n *g.Network, k, m int, opts ...g.Option) ([]*g.Switch, []*g.Host, error) {
	t := TopoOptions(opts...)

	hName := func(i, j int) string {
		return fmt.Sprintf("h%ds%d", j, i)
	}

	// if m == 1 {
	// 	hostName = func(i, j int) string {
	// 		return fmt.Sprintf("h%d", i)
	// 	}
	// }

	switches := []*g.Switch{}
	hosts := []*g.Host{}

	// var lastSwitch *g.Switch = nil

	for i := 0; i < k; i++ {
		// Add switch
		sw, err := n.AddSwitch(fmt.Sprintf("s%d", i), opts...)
		if err != nil {
			return nil, nil, err
		}

		// Add hosts to switch
		for j := 0; i < m; j++ {
			var h *g.Host

			intf := g.Interface{
				Port: g.Port{
					Name: "veth0",
					Node: sw,
				},
				Addresses: []net.IPNet{
					{
						IP: t.StartAddress,
					},
				},
			}

			if h, err = n.AddHost(hName(i, j), opts...); err != nil {
				return nil, nil, err
			}

			h.AddInterface(intf)

			hosts = append(hosts, h)
			t.StartAddress[2]++
		}

		// Link switch to previous
		// if lastSwitch != nil {
		// 	if err := t.LinkHosts(n, i, sw, lastSwitch, opts); err != nil {
		// 		return nil, nil, err
		// 	}
		// }

		// lastSwitch = sw
		switches = append(switches, sw)
		t.StartAddress[3]++
	}

	return switches, hosts, nil
}
