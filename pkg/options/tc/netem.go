// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package tc

import (
	"time"

	gopt "github.com/stv0g/gont/pkg/options"
)

type Latency time.Duration

func (m Latency) Apply(n *gopt.Netem) {
	d := time.Duration(m)
	n.Latency = uint32(d / time.Microsecond)
}

type Jitter time.Duration

func (j Jitter) Apply(n *gopt.Netem) {
	d := time.Duration(j)
	n.Jitter = uint32(d / time.Microsecond)
}

type Gap uint32

func (g Gap) Apply(n *gopt.Netem) {
	n.Gap = uint32(g)
}

type Loss Probability

func (p Loss) Apply(n *gopt.Netem) {
	n.Loss = p.Probability
	n.LossCorr = p.Correlation
}

type Reordering Probability

func (p Reordering) Apply(n *gopt.Netem) {
	n.ReorderProb = p.Probability
	n.ReorderCorr = p.Correlation
}

type Duplicate Probability

func (p Duplicate) Apply(n *gopt.Netem) {
	n.Duplicate = p.Probability
	n.DuplicateCorr = p.Correlation
}

type Corruption Probability

func (c Corruption) Apply(n *gopt.Netem) {
	n.CorruptProb = c.Probability
	n.CorruptCorr = c.Correlation
}

type LimitNetem uint32

func (l LimitNetem) Apply(n *gopt.Netem) {
	n.Limit = uint32(l)
}
