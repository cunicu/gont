package options

import (
	"net"

	g "github.com/stv0g/gont/pkg"
	nl "github.com/vishvananda/netlink"
)

// nl.LinkAttrs options

type MTU int

func (m MTU) Apply(la *nl.LinkAttrs) {
	la.MTU = int(m)
}

type Group g.DeviceGroup

func (h Group) Apply(p *g.Interface) {
	p.LinkAttrs.Group = uint32(h)
}

type TxQLen int

func (l TxQLen) Apply(la *nl.LinkAttrs) {
	la.TxQLen = int(l)
}

type HardwareAddress net.HardwareAddr

func (a HardwareAddress) Apply(la *nl.LinkAttrs) {
	la.HardwareAddr = net.HardwareAddr(a)
}

func AddressMAC(s string) HardwareAddress {
	mac, _ := net.ParseMAC(s)
	return HardwareAddress(mac)
}

func AddressMACBytes(b []byte) HardwareAddress {
	return HardwareAddress(b)
}
