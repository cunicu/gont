// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	g "cunicu.li/gont/v2/pkg"
)

// Persistent keeps a network from being torn down.
type Persistent bool

func (p Persistent) ApplyNetwork(n *g.Network) {
	n.Persistent = bool(p)
}

// Disable IPv6.
type IPv6Disabled bool

func (d IPv6Disabled) ApplyNetwork(n *g.Network) {
	n.IPv6Disabled = bool(d)
}
