package options

import (
	"net"

	nl "github.com/vishvananda/netlink"
)

// nl.LinkAttrs options

type MTU int
type HardwareAddress net.HardwareAddr

func (m MTU) Apply(la *nl.LinkAttrs) {
	la.MTU = int(m)
}

func (a HardwareAddress) Apply(la *nl.LinkAttrs) {
	la.HardwareAddr = net.HardwareAddr(a)
}
