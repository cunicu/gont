package options

import (
	g "github.com/stv0g/gont/pkg"
	nl "github.com/vishvananda/netlink"
)

type Netem nl.Netem
type Tbf nl.Tbf

func (ne *Netem) Apply(l *g.Link) {
	l.Netem = (*nl.Netem)(ne)
}

func (tbf *Tbf) Apply(l *g.Link) {
	l.Tbf = (*nl.Tbf)(tbf)
}
