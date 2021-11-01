package options

import (
	"net"

	g "github.com/stv0g/gont/pkg"
)

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
