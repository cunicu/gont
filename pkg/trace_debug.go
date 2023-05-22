// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build linux && (amd64 || arm64 || 386)

package gont

// Limit debugger support to platforms which are supported by Delve
// See: https://github.com/go-delve/delve/blob/master/pkg/proc/native/support_sentinel_linux.go

func (t *Tracer) ApplyDebugger(d *Debugger) {
	d.Tracers = append(d.Tracers, t)
}
