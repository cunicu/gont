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

type CaptureNetwork struct {
	*g.Capture
}

func (c CaptureNetwork) Apply(n *g.Network) {
	n.Captures = append(n.Captures, c.Capture)
}

func CaptureAll(opts ...g.CaptureOption) CaptureNetwork {
	return CaptureNetwork{Capture(opts...)}
}

type TraceNetwork struct {
	*g.Tracer
}

func (t TraceNetwork) Apply(n *g.Network) {
	n.Tracer = t.Tracer
}

func TraceAll(opts ...g.TraceOption) TraceNetwork {
	return TraceNetwork{Trace(opts...)}
}
