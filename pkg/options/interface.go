// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"fmt"
	"net"

	g "cunicu.li/gont/v2/pkg"
)

// Address assigns an network address (IP layer) to the interface.
type Address net.IPNet

func (a Address) ApplyInterface(i *g.Interface) {
	i.Addresses = append(i.Addresses, net.IPNet(a))
}

func AddressIPv4(a, b, c, d byte, m int) Address {
	return Address{
		IP:   net.IPv4(a, b, c, d),
		Mask: net.CIDRMask(m, 32),
	}
}

func AddressIP(fmts string, args ...any) Address {
	str := fmt.Sprintf(fmts, args...)

	ip, n, err := net.ParseCIDR(str)
	if err != nil {
		panic(fmt.Errorf("failed to parse IP address '%s': %w", str, err))
	}

	return Address{
		IP:   ip,
		Mask: n.Mask,
	}
}

// Disable duplicate address detection (DAD) for the interface.
type DADDisabled bool

func (d DADDisabled) ApplyInterface(i *g.Interface) {
	i.DADDisabled = bool(d)
}
