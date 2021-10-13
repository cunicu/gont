package gont

import (
	"fmt"
	"net"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type Host struct {
	BaseNode

	// Options
	GatewayIPv4 net.IP
	GatewayIPv6 net.IP
	Forwarding  bool

	Interfaces []Interface
}

// Getter
func (h *Host) Base() *BaseNode {
	return &h.BaseNode
}

// Options
func (h *Host) Apply(i *Interface) {
	i.Node = h
}

func (n *Network) AddHost(name string, opts ...Option) (*Host, error) {
	node, err := n.AddNode(name, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %s", err)
	}

	host := &Host{
		BaseNode:   *node,
		Interfaces: []Interface{},
	}

	n.Nodes[name] = host // TODO: quirk to get n.UpdateHostsFile() working

	// Apply host options
	for _, opt := range opts {
		if hopt, ok := opt.(HostOption); ok {
			hopt.Apply(host)
		}
	}

	// Add interfaces
	for _, intf := range host.Interfaces {
		if err := host.AddInterface(intf); err != nil {
			return nil, fmt.Errorf("failed to add interface: %w", err)
		}
	}

	// Configure host
	if host.GatewayIPv4 != nil {
		host.AddRoute(net.IPNet{
			IP:   net.IPv4zero,
			Mask: net.CIDRMask(0, net.IPv4len*8),
		}, host.GatewayIPv4)
	}

	if host.GatewayIPv6 != nil {
		host.AddRoute(net.IPNet{
			IP:   net.IPv6zero,
			Mask: net.CIDRMask(0, net.IPv6len*8),
		}, host.GatewayIPv6)
	}

	if host.Forwarding {
		if err := host.EnableForwarding(); err != nil {
			return nil, fmt.Errorf("failed to enable forwarding: %w", err)
		}
	}

	return host, nil
}

func (h *Host) AddInterface(i Interface) error {
	peerDev := fmt.Sprintf("veth-%s", h.name)

	l := Interface{
		Port: Port{
			Name: i.Name,
			Node: h,
		},
		Addresses: i.Addresses,
	}

	r := Port{
		Name: peerDev,
		Node: i.Node,
	}

	log.WithFields(log.Fields{
		"intf":      l,
		"intf_peer": r,
		"addresses": l.Addresses,
	}).Info("Adding interface")

	return h.Network.AddLink(l, r)
}

func (h *Host) ConfigureInterface(i Interface) error {
	// Disable duplicate address detection (DAD) before adding addresses
	// so we dont end up with tentative addresses and slow test executions
	if !i.EnableDAD {
		fn := filepath.Join("/proc/sys/net/ipv6/conf", i.Name, "accept_dad")
		if err := h.WriteProcFS(fn, "0"); err != nil {
			return err
		}
	}

	for _, addr := range i.Addresses {
		if err := h.LinkAddAddress(i.Name, addr); err != nil {
		}
	}

	// Bring interface up
	if err := h.BaseNode.ConfigurePort(i.Port); err != nil {
		return err
	}

	h.Interfaces = append(h.Interfaces, i) // TODO: arent the interface already in there for some cases?

	if err := h.Network.UpdateHostsFile(); err != nil {
		return err
	}

	return nil
}

func (h *Host) Ping(o *Host, opts ...string) error {
	arg := append([]string{"-c", "1", o.name}, opts...)
	_, _, err := h.Run("ping", arg...)
	return err
}

func (h *Host) Traceroute(o *Host) error {
	_, _, err := h.Run("traceroute", o.name)
	return err
}
