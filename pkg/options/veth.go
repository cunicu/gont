package options

import (
	"net"

	nl "github.com/vishvananda/netlink"
)

type PeerHardwareAddress net.HardwareAddr

func (p PeerHardwareAddress) Apply(v *nl.Veth) {
	v.PeerHardwareAddr = net.HardwareAddr(p)
}
