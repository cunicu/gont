// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"net"

	g "github.com/stv0g/gont/pkg"
	nl "github.com/vishvananda/netlink"
)

// nl.LinkAttrs options

type MTU int

func (m MTU) ApplyLink(la *nl.LinkAttrs) {
	la.MTU = int(m)
}

type Group g.DeviceGroup

func (h Group) ApplyLink(la *nl.LinkAttrs) {
	la.Group = uint32(g.DeviceGroup(h))
}

type TxQLen int

func (l TxQLen) ApplyLink(la *nl.LinkAttrs) {
	la.TxQLen = int(l)
}

type HardwareAddress net.HardwareAddr

func (a HardwareAddress) ApplyLink(la *nl.LinkAttrs) {
	la.HardwareAddr = net.HardwareAddr(a)
}

func AddressMAC(s string) HardwareAddress {
	mac, _ := net.ParseMAC(s)
	return HardwareAddress(mac)
}

func AddressMACBytes(b []byte) HardwareAddress {
	return HardwareAddress(b)
}
