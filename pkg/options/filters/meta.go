// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package filters

import (
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
)

func cstring(n string) []byte {
	b := make([]byte, 16)
	copy(b, []byte(n+"\x00"))
	return b
}

func interfaceName(dir expr.MetaKey, name string) Statement {
	return Statement{
		&expr.Meta{
			Key:      dir,
			Register: 1,
		},
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     cstring(name),
		},
	}
}

func interfaceIndexGroup(dir expr.MetaKey, idx uint32) Statement {
	return Statement{
		&expr.Meta{
			Key:      dir,
			Register: 1,
		},
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     binaryutil.BigEndian.PutUint32(idx),
		},
	}
}

func OutputInterfaceName(name string) Statement {
	return interfaceName(expr.MetaKeyOIFNAME, name)
}

func InputInterfaceName(name string) Statement {
	return interfaceName(expr.MetaKeyIIFNAME, name)
}

func OutputInterfaceIndex(idx uint32) Statement {
	return interfaceIndexGroup(expr.MetaKeyOIF, idx)
}

func InputInterfaceIndex(idx uint32) Statement {
	return interfaceIndexGroup(expr.MetaKeyIIF, idx)
}

func OutputInterfaceGroup(idx uint32) Statement {
	return interfaceIndexGroup(expr.MetaKeyOIFGROUP, idx)
}

func InputInterfaceGroup(idx uint32) Statement {
	return interfaceIndexGroup(expr.MetaKeyIIFGROUP, idx)
}

func Protocol(proto int) Statement {
	return Statement{
		&expr.Meta{
			Key:      expr.MetaKeyNFPROTO,
			Register: 1,
		},
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     binaryutil.NativeEndian.PutUint32(uint32(proto)),
		},
	}
}

func TransportProtocol(proto int) Statement {
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
