// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package cgroup

import (
	g "cunicu.li/gont/v2/pkg"
	sdbus "github.com/coreos/go-systemd/v22/dbus"
)

// Property is configuring a systemd CGroup resource control property
//
// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html
// See implementation:
//   - https://github.com/systemd/systemd/blob/e127c66985b9e338fcb79900a88a24848d5d31fb/src/core/dbus-cgroup.c
//   - https://github.com/systemd/systemd/blob/e127c66985b9e338fcb79900a88a24848d5d31fb/src/core/dbus-scope.c
//   - https://github.com/systemd/systemd/blob/e127c66985b9e338fcb79900a88a24848d5d31fb/src/core/dbus-slice.c
type Property sdbus.Property

func (p Property) ApplyCGroup(g *g.CGroup) {
	g.Properties = append(g.Properties, sdbus.Property(p))
}
