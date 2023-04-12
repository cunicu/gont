// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package filters

import (
	"fmt"
	"net"

	"github.com/google/nftables/expr"
	"github.com/stv0g/gont/v2/internal/utils"
)

type direction int

const (
	dirSource direction = iota
	dirDestination
)

func ipOffsetLen(ip net.IP, dir direction) (uint32, uint32) {
	isV4 := ip.To4() != nil

	switch {
	case isV4 && dir == dirSource:
		return 12, net.IPv4len
	case isV4 && dir == dirDestination:
		return 16, net.IPv4len
	case !isV4 && dir == dirSource:
		return 8, net.IPv6len
	case !isV4 && dir == dirDestination:
		return 24, net.IPv6len
	default:
		return 0, 0
	}
}

func network(dir direction, netw *net.IPNet) Statement {
	offset, len := ipOffsetLen(netw.IP, dir)
	fromAddr, toAddr := utils.AddressRange(netw)

	return Statement{
		// Source Address
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       offset,
			Len:          len,
		},
		&expr.Range{
			Op:       expr.CmpOpEq,
			Register: 1,
			FromData: fromAddr,
			ToData:   toAddr,
		},
	}
}

func Source(netw *net.IPNet) Statement {
	return network(dirSource, netw)
}

func Destination(netw *net.IPNet) Statement {
	return network(dirDestination, netw)
}

func SourceIP(fmts string, args ...any) Statement {
	str := fmt.Sprintf(fmts, args...)

	_, netw, err := net.ParseCIDR(str)
	if err != nil {
		panic(fmt.Errorf("failed to parse CIDR: %w", err))
	}

	return Source(netw)
}

func DestinationIP(fmts string, args ...any) Statement {
	str := fmt.Sprintf(fmts, args...)

	_, netw, err := net.ParseCIDR(str)
	if err != nil {
		panic(fmt.Errorf("failed to parse CIDR: %w", err))
	}

	return Destination(netw)
}

func SourceIPv4(a, b, c, d byte, m int) Statement {
	return Source(&net.IPNet{
		IP:   net.IPv4(a, b, c, d),
		Mask: net.CIDRMask(m, 32),
	})
}

func DestinationIPv4(a, b, c, d byte, m int) Statement {
	return Destination(&net.IPNet{
		IP:   net.IPv4(a, b, c, d),
		Mask: net.CIDRMask(m, 32),
	})
}
