package topo

import g "github.com/stv0g/gont/pkg"

func CreateHostSimple(n *g.Network, coord Vector, opts ...g.Option) (*g.Host, error) {
	name := "h"

	return n.AddHost(name, opts...)
}

func LinkHostsDirect(n *g.Network, l, r *g.Host, opts ...g.Option) error {
	lIntf := g.Interface{}
	rIntf := g.Interface{}

	return n.AddLink(lIntf, rIntf, opts...)
}

func LinkHostsSwitched(n *g.Network, a, b *g.Host, opts ...g.Option) error {
	_, err := n.AddSwitch("sw1", opts...)
	if err != nil {
		return err
	}

	return nil
}

func LinkHostsRouted(n *g.Network, a, b *g.Host, opts ...g.Option) error {
	_, err := n.AddRouter("r1", opts...)
	if err != nil {
		return err
	}
	return nil
}

func LinkHostsNatted(n *g.Network, a, b *g.Host, opts ...g.Option) error {
	_, err := n.AddNAT("n1", opts...)
	if err != nil {
		return err
	}

	return nil
}
