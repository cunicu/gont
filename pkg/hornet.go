package gont

import (
	"net"
)

const (
	hostsFile = "/etc/hosts"
)

var loopbackDevice Interface = Interface{
	Name:    "lo",
	Address: net.IPv4(127, 0, 0, 1),
	Mask:    net.IPv4Mask(255, 0, 0, 0),
}

type Interface struct {
	Name    string
	Address net.IP
	Mask    net.IPMask
	Switch  *Switch
}

func (i *Interface) Network() *net.IPNet {
	return &net.IPNet{
		IP:   i.Address.Mask(i.Mask),
		Mask: i.Mask,
	}
}

func TestConnectivity(n1, n2 *Host) error {
	err := n1.Ping(n2)
	if err != nil {
		return err
	}

	err = n2.Ping(n1)
	if err != nil {
		return err
	}

	return nil
}
