// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import gont "cunicu.li/gont/v2/pkg"

// Redirect output of sub-processes to log
type RedirectToLog bool

func (l RedirectToLog) ApplyNetwork(n *gont.Network) {
	n.RedirectToLog = bool(l)
}

func (l RedirectToLog) ApplyBaseNode(n *gont.BaseNode) {
	n.RedirectToLog = bool(l)
}

func (l RedirectToLog) ApplyCmd(c *gont.Cmd) {
	c.RedirectToLog = bool(l)
}
