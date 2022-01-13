package gont

import (
	"net"

	nl "github.com/vishvananda/netlink"
)

var (
	DefaultIPv4Mask = net.IPNet{
		IP:   net.IPv4zero,
		Mask: net.CIDRMask(0, net.IPv4len*8),
	}

	DefaultIPv6Mask = net.IPNet{
		IP:   net.IPv6zero,
		Mask: net.CIDRMask(0, net.IPv6len*8),
	}
)

type Route struct {
	nl.Route
}

func (r Route) Apply(h *Host) {
	h.Routes = append(h.Routes, r.Route)
}
