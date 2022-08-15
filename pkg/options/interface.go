package options

import (
	"fmt"
	"net"

	g "github.com/stv0g/gont/pkg"
)

// Address assigns an network address (IP layer) to the interface.
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
	ip, n, err := net.ParseCIDR(str)
	if err != nil {
		panic(fmt.Errorf("failed to parse IP address '%s': %w", str, err))
	}

	return Address{
		IP:   ip,
		Mask: n.Mask,
	}
}

func Capture(opts ...g.Option) *g.Capture {
	c := g.NewCapture()

	for _, o := range opts {
		if o, ok := o.(g.CaptureOption); ok {
			o.Apply(c)
		}
	}

	return c
}
