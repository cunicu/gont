// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"math/big"
	"net"
)

func ipToInt(ip net.IP) (*big.Int, int) {
	val := &big.Int{}
	val.SetBytes([]byte(ip))

	switch {
	case len(ip) == net.IPv4len:
		return val, 32
	case len(ip) == net.IPv6len:
		return val, 128
	default:
		panic(fmt.Sprintf("unsupported address length %d", len(ip)))
	}
}

func intToIP(ipInt *big.Int, bits int) net.IP {
	ipBytes := ipInt.Bytes()
	ret := make([]byte, bits/8)
	// Pack our IP bytes into the end of the return array,
	// since big.Int.Bytes() removes front zero padding.
	for i := 1; i <= len(ipBytes); i++ {
		ret[len(ret)-i] = ipBytes[len(ipBytes)-i]
	}
	return net.IP(ret)
}

// AddressRange returns the first and last addresses in the given CIDR range.
func AddressRange(network *net.IPNet) (net.IP, net.IP) {
	// the first IP is easy
	firstIP := network.IP

	// the last IP is the network address OR NOT the mask address
	prefixLen, bits := network.Mask.Size()
	if prefixLen == bits {
		// Easy!
		// But make sure that our two slices are distinct, since they
		// would be in all other cases.
		lastIP := make([]byte, len(firstIP))
		copy(lastIP, firstIP)
		return firstIP, lastIP
	}

	firstIPInt, bits := ipToInt(firstIP)
	hostLen := uint(bits) - uint(prefixLen) //nolint:gosec
	lastIPInt := big.NewInt(1)
	lastIPInt.Lsh(lastIPInt, hostLen)
	lastIPInt.Add(lastIPInt, firstIPInt)

	return firstIP, intToIP(lastIPInt, bits)
}
