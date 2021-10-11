package options

import (
	"net"

	g "github.com/stv0g/gont/pkg"
	nl "github.com/vishvananda/netlink"
)

// nl.Veth options

type PeerHardwareAddress net.HardwareAddr

func (p PeerHardwareAddress) apply(v *nl.Veth) {
	v.PeerHardwareAddr = net.HardwareAddr(p)
}

// nl.Bridge options

type BridgeOption interface {
	g.Option
	apply(b *nl.Bridge)
}
