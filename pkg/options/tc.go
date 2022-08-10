package options

import (
	g "github.com/stv0g/gont/pkg"
	nl "github.com/vishvananda/netlink"
)

// Network emulation

type NetemOption interface {
	Apply(n *Netem)
}

type Netem nl.NetemQdiscAttrs

func WithNetem(opts ...NetemOption) Netem {
	netem := Netem{}
	for _, opt := range opts {
		opt.Apply(&netem)
	}
	return netem
}

func (ne Netem) Apply(p *g.Interface) {
	p.Netem = nl.NetemQdiscAttrs(ne)
	p.Flags |= g.WithQdiscNetem
}

// Token Bucket Filter

type TbfOption interface {
	Apply(t *Tbf)
}

type Tbf nl.Tbf

func WithTbf(opts ...TbfOption) Tbf {
	tbf := Tbf{}
	for _, opt := range opts {
		opt.Apply(&tbf)
	}
	return tbf
}

func (tbf Tbf) Apply(p *g.Interface) {
	p.Tbf = nl.Tbf(tbf)
	p.Flags |= g.WithQdiscTbf
}
