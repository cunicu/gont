package options

import (
	"time"

	g "github.com/stv0g/gont/pkg"
	nl "github.com/vishvananda/netlink"
)

type Netem nl.NetemQdiscAttrs
type Tbf nl.Tbf

type NetemOption interface {
	Apply(n *Netem)
}

type TbfOption interface {
	Apply(t *Tbf)
}

func WithTbf(opts ...TbfOption) Tbf {
	tbf := Tbf{}
	for _, opt := range opts {
		opt.Apply(&tbf)
	}
	return tbf
}

func WithNetem(opts ...NetemOption) Netem {
	netem := Netem{}
	for _, opt := range opts {
		opt.Apply(&netem)
	}
	return netem
}

// General options

type Probability struct {
	Probability float32
	Correlation float32
}

func (ne Netem) Apply(l *g.Port) {
	l.Netem = nl.NetemQdiscAttrs(ne)
	l.Flags |= g.WithNetem
}

func (tbf Tbf) Apply(l *g.Port) {
	l.Tbf = nl.Tbf(tbf)
	l.Flags |= g.WithTbf
}

// Netem options

type Latency time.Duration

func (m Latency) Apply(n *Netem) {
	d := time.Duration(m)
	n.Latency = uint32(d / time.Microsecond)
}

type Jitter time.Duration

func (j Jitter) Apply(n *Netem) {
	d := time.Duration(j)
	n.Jitter = uint32(d / time.Microsecond)
}

type Loss Probability

func (p Loss) Apply(n *Netem) {
	n.Loss = p.Probability
	n.LossCorr = p.Correlation
}

type Reorder Probability

func (p Reorder) Apply(n *Netem) {
	n.ReorderProb = p.Probability
	n.ReorderCorr = p.Correlation
}

type Duplicate Probability

func (p Duplicate) Apply(n *Netem) {
	n.Duplicate = p.Probability
	n.DuplicateCorr = p.Correlation
}

// Tbf options

type Rate int

func (r Rate) Apply(t *Tbf) {
	t.Rate = uint64(r)
}
