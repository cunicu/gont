package options

import (
	"net"

	g "github.com/stv0g/gont/pkg"
)

type Forwarding bool
type Gateway net.IP

func (b Forwarding) Apply(h *g.Host) {
	h.Forwarding = bool(b)
}

func (g Gateway) Apply(h *g.Host) {
	ip := net.IP(g)
	if ipv4 := ip.To4(); ipv4 != nil {
		h.GatewayIPv4 = ipv4
	} else if ipv6 := ip.To16(); ipv6 != nil {
		h.GatewayIPv6 = ipv6
	}
}

func GatewayIPv4(a, b, c, d byte) Gateway {
	return Gateway(
		net.IPv4(a, b, c, d),
	)
}

func GatewayIP(str string) Gateway {
	return Gateway(
		net.ParseIP(str),
	)
}
