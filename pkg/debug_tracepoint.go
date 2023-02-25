// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"github.com/go-delve/delve/service/api"
)

type TracepointOption interface {
	ApplyTracepoint(*Tracepoint)
}

type Tracepoint struct {
	// Options
	api.Breakpoint
	Location string
	Message  string
}

func (bp Tracepoint) ApplyDebugger(d *Debugger) {
	d.Tracepoints = append(d.Tracepoints, bp)
}

func NewTracepoint(opts ...TracepointOption) Tracepoint {
	bp := Tracepoint{}

	for _, opt := range opts {
		opt.ApplyTracepoint(&bp)
	}

	return bp
}
