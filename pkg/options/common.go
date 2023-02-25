// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import g "github.com/stv0g/gont/pkg"

// Redirect output of sub-processes to log
type RedirectToLog bool

func (l RedirectToLog) ApplyNetwork(n *g.Network) {
	n.RedirectToLog = bool(l)
}

func (l RedirectToLog) ApplyBaseNode(n *g.BaseNode) {
	n.RedirectToLog = bool(l)
}

func (l RedirectToLog) ApplyCmd(c *g.Cmd) {
	c.RedirectToLog = bool(l)
}
