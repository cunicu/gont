package gont

import (
	"fmt"
	"net"

	nl "github.com/vishvananda/netlink"
)

const (
	WithQdiscNetem = (1 << iota)
	WithQdiscTbf   = (1 << iota)
)

var loopbackInterface = Interface{
	Name: loopbackInterfaceName,
	Addresses: []net.IPNet{
		{
			IP:   net.IPv4(127, 0, 0, 1),
			Mask: net.IPv4Mask(255, 0, 0, 0),
		},
		{
			IP:   net.IPv6loopback,
			Mask: net.CIDRMask(8*net.IPv6len, 8*net.IPv6len),
		},
	},
}

type InterfaceOption interface {
	Apply(n *Interface)
}

type Interface struct {
	Name string
	Node Node

	Link nl.Link

	Flags int

	// Options
	Netem     nl.NetemQdiscAttrs
	Tbf       nl.Tbf
	EnableDAD bool
	LinkAttrs nl.LinkAttrs
	Addresses []net.IPNet
	Captures  []*Capture
}

// Options

func (i *Interface) Apply(n *BaseNode) {
	n.ConfiguredInterfaces = append(n.ConfiguredInterfaces, i)
}

func (i Interface) String() string {
	if i.Node != nil {
		return fmt.Sprintf("%s/%s", i.Node, i.Name)
	}

	return i.Name
}

func (i Interface) IsLoopback() bool {
	return i.Name == loopbackInterfaceName
}

func (i *Interface) AddAddress(a *net.IPNet) error {
	return i.Node.NetlinkHandle().AddrAdd(i.Link, &nl.Addr{
		IPNet: a,
	})
}

func (i *Interface) DeleteAddress(a *net.IPNet) error {
	return i.Node.NetlinkHandle().AddrDel(i.Link, &nl.Addr{
		IPNet: a,
	})
}

func (i *Interface) SetUp() error {
	return i.Node.NetlinkHandle().LinkSetUp(i.Link)
}

func (i *Interface) SetDown() error {
	return i.Node.NetlinkHandle().LinkSetDown(i.Link)
}

func (i *Interface) Close() error {
	for _, c := range i.Captures {
		if err := c.Close(); err != nil {
			return fmt.Errorf("failed to close capture: %w", err)
		}
	}

	return nil
}
