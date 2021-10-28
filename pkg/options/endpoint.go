package options

import (
	"net"

	g "github.com/stv0g/gont/pkg"
)

type Group g.DeviceGroup

func (h Group) Apply(p *g.Port) {
	p.Group = g.DeviceGroup(h)
}

type Address net.IPNet

func (a Address) Apply(i *g.Interface) {
	i.Addresses = append(i.Addresses, net.IPNet(a))
}

func AddressIPv4(a, b, c, d byte, m int) Address {
	return Address{
		IP:   net.IPv4(a, b, c, d),
		Mask: net.CIDRMask(m, 32),
	}
}

func AddressIP(str string) Address {
	ip, n, _ := net.ParseCIDR(str)

	return Address{
		IP:   ip,
		Mask: n.Mask,
	}
}

func Port(name string, opts ...g.Option) g.Port {
	p := g.Port{
		Name: name,
	}

	// Apply port options
	for _, opt := range opts {
		if popt, ok := opt.(g.PortOption); ok {
			popt.Apply(&p)
		}
	}

	return p
}

func Interface(name string, opts ...g.Option) g.Interface {
	i := g.Interface{
		Port: Port(name, opts...),
	}

	for _, opt := range opts {
		if iopt, ok := opt.(g.InterfaceOption); ok {
			iopt.Apply(&i)
		}
	}

	return i
}
