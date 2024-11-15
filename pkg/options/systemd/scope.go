// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package cgroup

import (
	"time"

	sdbus "github.com/coreos/go-systemd/v22/dbus"
	"github.com/godbus/dbus/v5"
)

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html#RuntimeMaxSec=
func RuntimeMax(period time.Duration) Property {
	return Property(sdbus.Property{
		Name:  "RuntimeMaxUSec",
		Value: dbus.MakeVariant(uint64(period.Microseconds())), //nolint:gosec
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html#TimeoutStopSec=
func TimeoutStop(period time.Duration) Property {
	return Property(sdbus.Property{
		Name:  "TimeoutStopUSec",
		Value: dbus.MakeVariant(uint64(period.Microseconds())), //nolint:gosec
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html#RuntimeRandomizedExtraSec=
func RuntimeRandomizedExtra(period time.Duration) Property {
	return Property(sdbus.Property{
		Name:  "RuntimeRandomizedExtraUSec",
		Value: dbus.MakeVariant(uint64(period.Microseconds())), //nolint:gosec
	})
}

// See: https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html#OOMPolicy=
type OOMPolicyType string

const (
	OOMPolicyContinue OOMPolicyType = "continue"
	OOMPolicyStop     OOMPolicyType = "stop"
	OOMPolicyKill     OOMPolicyType = "kill"
)

func OOMPolicy(policy OOMPolicyType) Property {
	return Property(sdbus.Property{
		Name:  "OOMPolicy",
		Value: dbus.MakeVariant(policy),
	})
}

func PIDs(pids ...uint32) Property {
	return Property(sdbus.Property{
		Name:  "PIDs",
		Value: dbus.MakeVariant(pids),
	})
}

func PIDFDs(fds []int) Property {
	ufds := []dbus.UnixFD{}
	for _, fd := range fds {
		ufds = append(ufds, dbus.UnixFD(fd)) //nolint:gosec
	}

	return Property(sdbus.Property{
		Name:  "PIDFDs",
		Value: dbus.MakeVariant(ufds),
	})
}

func User(user string) Property {
	return Property(sdbus.Property{
		Name:  "User",
		Value: dbus.MakeVariant(user),
	})
}

func Group(group string) Property {
	return Property(sdbus.Property{
		Name:  "Group",
		Value: dbus.MakeVariant(group),
	})
}
