package filters

import (
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
)

func port(offset uint32, portNum uint16) Statement {
	return Statement{
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseTransportHeader,
			Offset:       offset,
			Len:          2,
		},
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     binaryutil.BigEndian.PutUint16(portNum),
		},
	}
}

func portRange(offset uint32, minPort, maxPort uint16) Statement {
	return Statement{
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseTransportHeader,
			Offset:       offset,
			Len:          2,
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

func SourcePort(portNum uint16) Statement {
	return port(0, portNum)
}

func DestinationPort(portNum uint16) Statement {
	return port(2, portNum)
}

func SourcePortRange(minPort, maxPort uint16) Statement {
	return portRange(0, minPort, maxPort)
}

func DestinationPortRange(minPort, maxPort uint16) Statement {
	return portRange(2, minPort, maxPort)
}
