package gont

import nl "github.com/vishvananda/netlink"

type Option any

type NodeOption interface {
	Apply(b *BaseNode)
}

type HostOption interface {
	Apply(h *Host)
}

type InterfaceOption interface {
	Apply(i *Interface)
}

type NATOption interface {
	Apply(n *NAT)
}

type SwitchOption interface {
	Apply(sw *Switch)
}

type NetworkOption interface {
	Apply(n *Network)
}

type VethOption interface {
	Apply(v *nl.Veth)
}

type LinkOption interface {
	Apply(la *nl.LinkAttrs)
}

type BridgeOption interface {
	Apply(b *nl.Bridge)
}

type CaptureOption interface {
	Apply(c *Capture)
}
