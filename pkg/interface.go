// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

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

//nolint:gochecknoglobals
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
	ApplyInterface(n *Interface)
}

func (i *Interface) ApplyBaseNode(n *BaseNode) {
	n.ConfiguredInterfaces = append(n.ConfiguredInterfaces, i)
}

type Interface struct {
	Name string
	node Node

	Link  nl.Link
	Flags int

	// Options
	Netem     nl.NetemQdiscAttrs
	Tbf       nl.Tbf
	EnableDAD bool
	LinkAttrs nl.LinkAttrs
	Addresses []net.IPNet
	Captures  []*Capture
}

func NewInterface(name string, opts ...Option) *Interface {
	i := &Interface{
		Name:     name,
		Captures: []*Capture{},
	}

	for _, o := range opts {
		switch opt := o.(type) {
		case InterfaceOption:
			opt.ApplyInterface(i)
		case LinkOption:
			opt.ApplyLink(&i.LinkAttrs)
		}
	}

	return i
}

func (i Interface) String() string {
	if i.node != nil {
		return fmt.Sprintf("%s/%s", i.node, i.Name)
	}

	return i.Name
}

func (i Interface) IsLoopback() bool {
	return i.Name == loopbackInterfaceName
}

func (i *Interface) AddAddress(a *net.IPNet) error {
	return i.node.NetlinkHandle().AddrAdd(i.Link, &nl.Addr{
		IPNet: a,
	})
}

func (i *Interface) DeleteAddress(a *net.IPNet) error {
	return i.node.NetlinkHandle().AddrDel(i.Link, &nl.Addr{
		IPNet: a,
	})
}

func (i *Interface) SetUp() error {
	return i.node.NetlinkHandle().LinkSetUp(i.Link)
}

func (i *Interface) SetDown() error {
	return i.node.NetlinkHandle().LinkSetDown(i.Link)
}

func (i *Interface) Close() error {
	for _, c := range i.Captures {
		if err := c.Close(); err != nil {
			return fmt.Errorf("failed to close capture: %w", err)
		}
	}

	return nil
}
