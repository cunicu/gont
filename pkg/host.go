package gont

import (
	"net"
)

type Host struct {
	Node

	Interface *Interface
}

func (n *Network) AddHost(name string, gw net.IP, intf *Interface) (*Host, error) {
	node, err := n.AddNode(name)
	if err != nil {
		return nil, err
	}

	host := &Host{
		Node:      *node,
		Interface: intf,
	}

	err = n.AddToHostsFile(intf.Address, name)
	if err != nil {
		return nil, err
	}

	if _, _, err = host.Run("ip", "link", "add", "name", host.Interface.Name, "type", "veth", "peer", "br0-"+host.Name+"-"+intf.Name, "netns", intf.Switch.NS.String()); err != nil {
		return nil, err
	}
	if _, _, err = intf.Switch.Run("ip", "link", "set", "dev", "br0-"+host.Name+"-"+host.Interface.Name, "master", "br0"); err != nil {
		return nil, err
	}
	if _, _, err = intf.Switch.Run("ip", "link", "set", "dev", "br0-"+host.Name+"-"+host.Interface.Name, "up"); err != nil {
		return nil, err
	}

	host.ConfigureInterface(intf)
	host.ConfigureInterface(&loopbackDevice)

	if gw != nil {
		if _, _, err = host.Run("ip", "route", "add", "default", "via", gw.String()); err != nil {
			return nil, err
		}
	}

	return host, nil
}

func (h *Host) Close(o *Host) error {
	return h.Network.RemoveFromHostsFile(h.Interface.Address)
}

func (h *Host) Ping(o *Host, opts ...string) error {
	cmd := append([]string{"ping", o.Name}, opts...)
	_, _, err := h.Run(cmd...)
	return err
}

func (h *Host) Traceroute(o *Host) error {
	_, _, err := h.Run("traceroute", o.Name)
	return err
}
