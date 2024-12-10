// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package cgroup

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/godbus/dbus/v5"
	"golang.org/x/sys/unix"
)

var (
	errNetworkUnsupported = errors.New("unsupported network")
	errInvalidPorts       = errors.New("invalid ports")
)

func makeCPUSet(mask uint64) dbus.Variant {
	buf := binary.LittleEndian.AppendUint64(nil, mask)
	return dbus.MakeVariant(buf)
}

func makePrefixList(prefixes []net.IPNet) dbus.Variant {
	type prefix struct {
		Family int    `dbus:"-"`
		Prefix []byte `dbus:"-"`
		Len    uint32 `dbus:"-"`
	}

	ps := []prefix{}
	for _, pfx := range prefixes {
		l, _ := pfx.Mask.Size()

		p := prefix{
			Len: uint32(l), //nolint:gosec
		}

		if b := pfx.IP.To4(); b == nil {
			p.Family = unix.AF_INET6
			p.Prefix = b
		} else {
			p.Family = unix.AF_INET
			p.Prefix = b
		}

		ps = append(ps, p)
	}

	return dbus.MakeVariant(ps)
}

type BindPolicy struct {
	// Network family. Must be one of "tcp4", "tcp6", "udp4", "udp6"
	Network string
	PortMin uint16
	PortMax uint16
}

func makeBindPolicies(policies []BindPolicy) (dbus.Variant, error) {
	type policy struct {
		Family  int    `dbus:"-"`
		IPProto int    `dbus:"-"`
		PortMin uint16 `dbus:"-"`
		NrPorts uint16 `dbus:"-"`
	}

	ps := []policy{}
	for _, pol := range policies {
		if pol.PortMin > pol.PortMax {
			return dbus.MakeVariant(nil), fmt.Errorf("%w: %d > %d", errInvalidPorts, pol.PortMin, pol.PortMax)
		}

		p := policy{
			PortMin: pol.PortMin,
			NrPorts: pol.PortMax - pol.PortMin + 1,
		}

		switch pol.Network {
		case "tcp4":
			p.Family = unix.AF_INET
			p.IPProto = unix.IPPROTO_TCP
		case "tcp6":
			p.Family = unix.AF_INET6
			p.IPProto = unix.IPPROTO_TCP
		case "udp4":
			p.Family = unix.AF_INET
			p.IPProto = unix.IPPROTO_UDP
		case "udp6":
			p.Family = unix.AF_INET6
			p.IPProto = unix.IPPROTO_UDP
		default:
			return dbus.MakeVariant(nil), fmt.Errorf("%w: %s", errNetworkUnsupported, pol.Network)
		}

		ps = append(ps, p)
	}

	return dbus.MakeVariant(ps), nil
}

func makeNetworkInterfaceAllowList(allow bool, intfs []string) dbus.Variant {
	type allowList struct {
		Allow          bool     `dbus:"-"`
		InterfaceNames []string `dbus:"-"`
	}

	return dbus.MakeVariant(allowList{
		Allow:          allow,
		InterfaceNames: intfs,
	})
}
