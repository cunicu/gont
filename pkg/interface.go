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

var loopbackInterface Interface = Interface{
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
}

// Options

func (i *Interface) Apply(n *BaseNode) {
	n.ConfiguredInterfaces = append(n.ConfiguredInterfaces, i)
}

func (i Interface) String() string {
	if i.Node != nil {
		return fmt.Sprintf("%s/%s", i.Node, i.Name)
	} else {
		return i.Name
	}
}

func (i Interface) IsLoopback() bool {
	return i.Name == loopbackInterfaceName
}

func (i *Interface) Configure() error {
	return i.Node.ConfigureInterface(i)
}
