// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Limit debugger support to platforms which are supported by Delve
// See: https://github.com/go-delve/delve/blob/master/pkg/proc/native/support_sentinel_linux.go
//go:build linux && (amd64 || arm64 || 386)

package debug

import "github.com/go-delve/delve/service/api"

// LoadConfigOption is an interface implemented by all options
// which modify Delves api.LoadConfig to configure which information
// is loaded when a breakpoint is hit.
type LoadConfigOption interface {
	ApplyLoadConfig(lc *api.LoadConfig)
}

// FollowPointers requests pointers to be automatically dereferenced.
type FollowPointers bool

func (f FollowPointers) ApplyLoadConfig(lc *api.LoadConfig) {
	lc.FollowPointers = bool(f)
}

// MaxVariableRecurse is how far to recurse when evaluating nested types.
type MaxVariableRecurse int

func (m MaxVariableRecurse) ApplyLoadConfig(lc *api.LoadConfig) {
	lc.MaxVariableRecurse = int(m)
}

// MaxStringLen is the maximum number of bytes read from a string
type MaxStringLen int

func (m MaxStringLen) ApplyLoadConfig(lc *api.LoadConfig) {
	lc.MaxStringLen = int(m)
}

// MaxArrayValues is the maximum number of elements read from an array, a slice or a map.
type MaxArrayValues int

func (m MaxArrayValues) ApplyLoadConfig(lc *api.LoadConfig) {
	lc.MaxArrayValues = int(m)
}

// MaxStructFields is the maximum number of fields read from a struct, -1 will read all fields.
type MaxStructFields int

func (m MaxStructFields) ApplyLoadConfig(lc *api.LoadConfig) {
	lc.MaxStructFields = int(m)
}
