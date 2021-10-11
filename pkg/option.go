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

type PortOption interface {
	Option
	Apply(p *Port)
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

type LinkOption interface {
	Option
	Apply(l *Link)
}

type NetworkOption interface {
	Option
	Apply(n *Network)
}

type VethOption interface {
	Option
	Apply(v *nl.Veth)
}

type LinkAttrOption interface {
	Option
	apply(la *nl.LinkAttrs)
}
