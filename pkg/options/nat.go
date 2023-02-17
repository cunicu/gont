// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import g "github.com/stv0g/gont/pkg"

const (
	SouthBound = Group(g.DeviceGroupSouthBound)
	NorthBound = Group(g.DeviceGroupNorthBound)
)

type PersistentNAT bool

func (pn PersistentNAT) Apply(n *g.NAT) {
	n.Persistent = bool(pn)
}

type RandomNAT bool

func (rn RandomNAT) Apply(n *g.NAT) {
	n.Random = bool(rn)
}

type FullyRandomNAT bool

func (frn FullyRandomNAT) Apply(n *g.NAT) {
	n.FullyRandom = bool(frn)
}

type SourcePortRange struct {
	Min int
	Max int
}

func (spr SourcePortRange) Apply(n *g.NAT) {
	n.SourcePortMin = spr.Min
	n.SourcePortMax = spr.Max
}
