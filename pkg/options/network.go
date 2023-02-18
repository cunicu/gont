// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	g "github.com/stv0g/gont/pkg"
)

type NSPrefix string

func (pfx NSPrefix) Apply(n *g.Network) {
	n.NSPrefix = string(pfx)
}

// Persistent keeps a network from beeing torn down.
type Persistent bool

func (p Persistent) Apply(n *g.Network) {
	n.Persistent = bool(p)
}
