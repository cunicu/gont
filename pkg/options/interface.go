// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"fmt"
	"net"

	g "github.com/stv0g/gont/pkg"
)

// Address assigns an network address (IP layer) to the interface.
type Address net.IPNet

func (a Address) Apply(i *g.Interface) {
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

func Capture(opts ...g.CaptureOption) *g.Capture {
	c := g.NewCapture()

	for _, o := range opts {
		o.Apply(c)
	}

	return c
}
