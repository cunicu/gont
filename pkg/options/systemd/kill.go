// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package cgroup

import (
	sdbus "github.com/coreos/go-systemd/v22/dbus"
	"github.com/godbus/dbus/v5"
)

// See: https://github.com/systemd/systemd/blob/main/src/core/dbus-kill.c

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.kill.html#KillMode=
type KillModeType string

const (
	KillModeControlGroup KillModeType = "control-group"
	KillModeProcess      KillModeType = "process"
	KillModeMixed        KillModeType = "mixed"
	KillModeNonde        KillModeType = "none"
)

func KillMode(typ KillModeType) Property {
	return Property(sdbus.Property{
		Name:  "KillMode",
		Value: dbus.MakeVariant(typ),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.kill.html#KillSignal=
func KillSignal(signal int) Property {
	return Property(sdbus.Property{
		Name:  "KillSignal",
		Value: dbus.MakeVariant(signal),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.kill.html#RestartKillSignal=
func RestartKillSignal(signal int) Property {
	return Property(sdbus.Property{
		Name:  "RestartKillSignal",
		Value: dbus.MakeVariant(signal),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.kill.html#SendSIGHUP=
func SendSIGHUP(enable bool) Property {
	return Property(sdbus.Property{
		Name:  "SendSIGHUP",
		Value: dbus.MakeVariant(enable),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.kill.html#SendSIGKILL=
func SendSIGKILL(enable bool) Property {
	return Property(sdbus.Property{
		Name:  "SendSIGKILL",
		Value: dbus.MakeVariant(enable),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.kill.html#FinalKillSignal=
func FinalKillSignal(signal int) Property {
	return Property(sdbus.Property{
		Name:  "FinalKillSignal",
		Value: dbus.MakeVariant(signal),
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.kill.html#WatchdogSignal=
func WatchdogSignal(signal int) Property {
	return Property(sdbus.Property{
		Name:  "WatchdogSignal",
		Value: dbus.MakeVariant(signal),
	})
}
