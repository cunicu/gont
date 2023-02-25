// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package tc

import o "github.com/stv0g/gont/pkg/options"

type Rate uint64

func (r Rate) ApplyTbf(t *o.Tbf) {
	t.Rate = uint64(r)
}

type Buffer uint32

func (r Buffer) ApplyTbf(t *o.Tbf) {
	t.Buffer = uint32(r)
}

type PeakRate uint64

func (r PeakRate) ApplyTbf(t *o.Tbf) {
	t.Peakrate = uint64(r)
}

type MinBurst uint32

func (r MinBurst) ApplyTbf(t *o.Tbf) {
	t.Minburst = uint32(r)
}

type LimitTbf uint32

func (l LimitTbf) ApplyTbf(t *o.Tbf) {
	t.Limit = uint32(l)
}
