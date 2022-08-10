package tc

import gopt "github.com/stv0g/gont/pkg/options"

type Rate uint64

func (r Rate) Apply(t *gopt.Tbf) {
	t.Rate = uint64(r)
}

type Buffer uint32

func (r Buffer) Apply(t *gopt.Tbf) {
	t.Buffer = uint32(r)
}

type PeakRate uint64

func (r PeakRate) Apply(t *gopt.Tbf) {
	t.Peakrate = uint64(r)
}

type MinBurst uint32

func (r MinBurst) Apply(t *gopt.Tbf) {
	t.Minburst = uint32(r)
}

type LimitTbf uint32

func (l LimitTbf) Apply(t *gopt.Tbf) {
	t.Limit = uint32(l)
}
