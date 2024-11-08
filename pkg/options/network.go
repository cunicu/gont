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
