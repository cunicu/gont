// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	g "cunicu.li/gont/v2/pkg"
)

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

// Slice set the systemd slice / CGroup in which all processes of this node, network or command are started.
type Slice string

func (e Slice) ApplyBaseNode(n *g.BaseNode) {
	n.Slice = string(e)
}

func (e Slice) ApplyNetwork(n *g.Network) {
	n.Slice = string(e)
}

func (e Slice) ApplyCmd(c *g.Cmd) {
	c.Slice = string(e)
}
