// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	g "github.com/stv0g/gont/pkg"
)

type NSPrefix string

func (pfx NSPrefix) ApplyNetwork(n *g.Network) {
	n.NSPrefix = string(pfx)
}

// Persistent keeps a network from being torn down.
type Persistent bool

func (p Persistent) ApplyNetwork(n *g.Network) {
	n.Persistent = bool(p)
}
