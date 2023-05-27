// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Limit debugger support to platforms which are supported by Delve
// See: https://github.com/go-delve/delve/blob/master/pkg/proc/native/support_sentinel_linux.go
//go:build !(linux && (amd64 || arm64 || 386))

package trace

import "io"

type (
	Breakpoint struct{}
)

func (*Breakpoint) Fprint(io.Writer, string) {}
