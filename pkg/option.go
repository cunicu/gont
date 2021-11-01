package gont

import nl "github.com/vishvananda/netlink"

type Option interface{}
type Options []Option

type NodeOption interface {
	Option
	Apply(b *BaseNode)
}

type HostOption interface {
	Option
	Apply(h *Host)
}

type InterfaceOption interface {
	Option
	Apply(i *Interface)
}

type NATOption interface {
	Option
	Apply(n *NAT)
}

type SwitchOption interface {
	Option
	Apply(sw *Switch)
}

type NetworkOption interface {
	Option
	Apply(n *Network)
}

type VethOption interface {
	Option
	Apply(v *nl.Veth)
}

type LinkOption interface {
	Option
	Apply(la *nl.LinkAttrs)
}

type BridgeOption interface {
	Apply(b *nl.Bridge)
}
