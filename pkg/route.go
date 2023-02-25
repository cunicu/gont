// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"net"
)

//nolint:gochecknoglobals
var (
	DefaultIPv4Mask = net.IPNet{
		IP:   net.IPv4zero,
		Mask: net.CIDRMask(0, net.IPv4len*8),
	}

	DefaultIPv6Mask = net.IPNet{
		IP:   net.IPv6zero,
		Mask: net.CIDRMask(0, net.IPv6len*8),
	}
)
