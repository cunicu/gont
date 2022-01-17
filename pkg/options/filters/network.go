package filters

import (
	"net"

	"github.com/google/nftables/expr"
	"github.com/stv0g/gont/internal/utils"
)

type direction int

const (
	dirSource direction = iota
	dirDestination
)

func network(dir direction, netw *net.IPNet) Statement {
	var offset, len uint32

	isV4 := netw.IP.To4() != nil
	if isV4 {
		len = net.IPv4len
		if dir == dirSource {
			offset = 12
		} else {
			offset = 16
		}
	} else {
		len = net.IPv6len
		if dir == dirSource {
			offset = 8
		} else {
			offset = 24
		}
	}

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
