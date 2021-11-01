package options

import (
	"time"

	g "github.com/stv0g/gont/pkg"
	nl "github.com/vishvananda/netlink"
)

type Netem nl.NetemQdiscAttrs
type Tbf nl.Tbf

type NetemOption interface {
	ApplyNetem(n *Netem)
}

type TbfOption interface {
	ApplyTbf(t *Tbf)
}

func WithTbf(opts ...TbfOption) Tbf {
	tbf := Tbf{}
	for _, opt := range opts {
		opt.ApplyTbf(&tbf)
	}
	return tbf
}

func WithNetem(opts ...NetemOption) Netem {
	netem := Netem{}
	for _, opt := range opts {
		opt.ApplyNetem(&netem)
	}
	return netem
}

// General options

type Probability struct {
	Probability float32
	Correlation float32
}

func (ne Netem) Apply(p *g.Interface) {
	p.Netem = nl.NetemQdiscAttrs(ne)
	p.Flags |= g.WithQdiscNetem
}

func (tbf Tbf) Apply(p *g.Interface) {
	p.Tbf = nl.Tbf(tbf)
	p.Flags |= g.WithQdiscTbf
}

// Netem options

type Latency time.Duration

func (m Latency) ApplyNetem(n *Netem) {
	d := time.Duration(m)
	n.Latency = uint64(d / time.Microsecond)
}

type Jitter time.Duration

func (j Jitter) ApplyNetem(n *Netem) {
	d := time.Duration(j)
	n.Jitter = uint64(d / time.Microsecond)
}

type Gap uint32

func (g Gap) ApplyNetem(n *Netem) {
	n.Gap = uint32(g)
}

type Loss Probability

func (p Loss) ApplyNetem(n *Netem) {
	n.Loss = p.Probability
	n.LossCorr = p.Correlation
}

type Reordering Probability

func (p Reordering) ApplyNetem(n *Netem) {
	n.ReorderProb = p.Probability
	n.ReorderCorr = p.Correlation
}

type Duplicate Probability

func (p Duplicate) ApplyNetem(n *Netem) {
	n.Duplicate = p.Probability
	n.DuplicateCorr = p.Correlation
}

type Corruption Probability

func (c Corruption) ApplyNetem(n *Netem) {
	n.CorruptProb = c.Probability
	n.CorruptCorr = c.Correlation
}

// Tbf options

type Buffer uint32

func (r Buffer) Apply(t *Tbf) {
	t.Buffer = uint32(r)
}

type PeakRate uint64

func (r PeakRate) Apply(t *Tbf) {
	t.Peakrate = uint64(r)
}

type MinBurst uint32

func (r MinBurst) Apply(t *Tbf) {
	t.Minburst = uint32(r)
}

// Common options

type Limit uint32

func (l Limit) ApplyNetem(n *Netem) {
	n.Limit = uint32(l)
}

func (l Limit) ApplyTbf(t *Tbf) {
	t.Limit = uint32(l)
}

type Rate uint64

func (r Rate) ApplyNetem(n *Netem) {
	n.Rate = uint64(r)
}

func (r Rate) ApplyTbf(t *Tbf) {
	t.Rate = uint64(r)
}
