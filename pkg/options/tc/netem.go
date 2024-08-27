// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package tc

import (
	"time"

	o "cunicu.li/gont/v2/pkg/options"
)

type Latency time.Duration

func (m Latency) ApplyNetem(n *o.Netem) {
	d := time.Duration(m)
	n.Latency = uint32(d / time.Microsecond) //nolint:gosec
}

type Jitter time.Duration

func (j Jitter) ApplyNetem(n *o.Netem) {
	d := time.Duration(j)
	n.Jitter = uint32(d / time.Microsecond) //nolint:gosec
}

type Gap uint32

func (g Gap) ApplyNetem(n *o.Netem) {
	n.Gap = uint32(g)
}

type Loss Probability

func (p Loss) ApplyNetem(n *o.Netem) {
	n.Loss = p.Probability
	n.LossCorr = p.Correlation
}

type Reordering Probability

func (p Reordering) ApplyNetem(n *o.Netem) {
	n.ReorderProb = p.Probability
	n.ReorderCorr = p.Correlation
}

type Duplicate Probability

func (p Duplicate) ApplyNetem(n *o.Netem) {
	n.Duplicate = p.Probability
	n.DuplicateCorr = p.Correlation
}

type Corruption Probability

func (c Corruption) ApplyNetem(n *o.Netem) {
	n.CorruptProb = c.Probability
	n.CorruptCorr = c.Correlation
}

type LimitNetem uint32

func (l LimitNetem) ApplyNetem(n *o.Netem) {
	n.Limit = uint32(l)
}
