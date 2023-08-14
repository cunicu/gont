// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	g "cunicu.li/gont/v2/pkg"
	nl "github.com/vishvananda/netlink"
)

// Network emulation

type NetemOption interface {
	ApplyNetem(n *Netem)
}

type Netem nl.NetemQdiscAttrs

func WithNetem(opts ...NetemOption) Netem {
	netem := Netem{}

	for _, opt := range opts {
		opt.ApplyNetem(&netem)
	}

	return netem
}

func (ne Netem) ApplyInterface(p *g.Interface) {
	p.Netem = nl.NetemQdiscAttrs(ne)
	p.Flags |= g.WithQdiscNetem
}

// Token Bucket Filter

type TbfOption interface {
	ApplyTbf(t *Tbf)
}

type Tbf nl.Tbf

func WithTbf(opts ...TbfOption) Tbf {
	tbf := Tbf{}

	for _, opt := range opts {
		opt.ApplyTbf(&tbf)
	}

	return tbf
}

func (tbf Tbf) ApplyInterface(p *g.Interface) {
	p.Tbf = nl.Tbf(tbf)
	p.Flags |= g.WithQdiscTbf
}
