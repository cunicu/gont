package gont

import (
	"fmt"
	"net"
	"reflect"
)

var loopbackInterface Interface = Interface{
	Port: Port{
		Name: "lo",
	},
	Addresses: []net.IPNet{
		{
			IP:   net.IPv4(127, 0, 0, 1),
			Mask: net.IPv4Mask(255, 0, 0, 0),
		},
		{
			IP:   net.IPv6loopback,
			Mask: net.CIDRMask(0, 8*net.IPv6len),
		},
	},
}

type Endpoint interface {
	Configure() error
	BaseNode() *BaseNode
	port() Port
}

type Port struct {
	Endpoint

	Name string
	Node Node

	Group DeviceGroup
}

type Interface struct {
	Port
	Addresses []net.IPNet
}

// Getter

func (p Port) port() Port {
	return p
}

func (p Port) String() string {
	if p.Node != nil {
		return p.Node.Name() + "/" + p.Name
	} else {
		return p.Name
	}
}

func (i Interface) port() Port {
	return i.Port
}

func (i Interface) String() string {
	return i.Port.String()
}

// Options

func (i Interface) Apply(h *Host) {
	h.Interfaces = append(h.Interfaces, i)
}

func (p Port) Apply(sw *Switch) {
	sw.Ports = append(sw.Ports, p)
}

// Configure

func (p Port) Configure() error {
	if n, ok := p.Node.(*Switch); ok {
		return n.ConfigurePort(p)
	}

	return fmt.Errorf("cant configure ports for nodes of type %s", reflect.TypeOf(p.Node).String())
}

func (i Interface) Configure() error {
	if h, ok := i.Port.Node.(*Host); ok {
		return h.ConfigureInterface(i)
	} else if r, ok := i.Port.Node.(*Router); ok {
		return r.ConfigureInterface(i)
	}

	return fmt.Errorf("cant configure interface for %s", reflect.TypeOf(i.Port.Node).String())
}
