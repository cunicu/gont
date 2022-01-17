package filters

import (
	"net"

	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"github.com/stv0g/gont/internal/utils"
)

// Statement is a list of one or more nftables expressions
type Statement []expr.Any

var Drop = Statement{
	&expr.Verdict{
		Kind: expr.VerdictDrop,
	},
}

func Source(src *net.IPNet) Statement {
	var offset, len uint32

	isV4 := src.IP.To4() != nil
	if isV4 {
		offset = 12
		len = net.IPv4len
	} else {
		offset = 8
		len = net.IPv6len
	}

	fromAddr, toAddr := utils.AddressRange(src)

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

func Protocol(proto int) Statement {
	return Statement{
		// Protocol
		&expr.Meta{
			Key:      expr.MetaKeyL4PROTO,
			Register: 1,
		},
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     []byte{byte(proto)},
		},
	}
}

func Port(port uint16) Statement {
	return Statement{
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseTransportHeader,
			Offset:       2, // TODO
			Len:          2, // TODO
		},
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     binaryutil.BigEndian.PutUint16(port),
		},
	}
}

func PortRange(minPort, maxPort uint16) Statement {
	return Statement{
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseTransportHeader,
			Offset:       2, // TODO
			Len:          2, // TODO
		},
		&expr.Cmp{
			Op:       expr.CmpOpGte,
			Register: 1,
			Data:     binaryutil.BigEndian.PutUint16(minPort),
		},
		&expr.Cmp{
			Op:       expr.CmpOpLte,
			Register: 1,
			Data:     binaryutil.BigEndian.PutUint16(maxPort),
		},
	}
}
