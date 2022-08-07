package gont

import (
	"fmt"
	"net"
	"path/filepath"

	nl "github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

type Host struct {
	*BaseNode

	Filter *Filter

	// Options
	FilterRules []*FilterRule
	Routes      []*nl.Route
	Forwarding  bool
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
		BaseNode:    node,
		Routes:      []*nl.Route{},
		FilterRules: []*FilterRule{},
	}

	n.Register(host)

	// Apply host options
	for _, opt := range opts {
		if hopt, ok := opt.(HostOption); ok {
			hopt.Apply(host)
		}
	}

	// Configure loopback device
	lo := loopbackInterface
	lo.Node = host
	if lo.Link, err = host.nlHandle.LinkByName("lo"); err != nil {
		return nil, fmt.Errorf("failed to get loopback interface: %w", err)
	}
	if err := host.ConfigureInterface(&lo); err != nil {
		return nil, fmt.Errorf("failed to configure loopback interface: %w", err)
	}

	if err := host.ConfigureLinks(); err != nil {
		return nil, fmt.Errorf("failed to configure links: %w", err)
	}

	// Configure host
	for _, r := range host.Routes {
		if err := host.AddRoute(r); err != nil {
			return nil, fmt.Errorf("failed to add route: %w", err)
		}
	}

	if host.Forwarding {
		if err := host.EnableForwarding(); err != nil {
			return nil, fmt.Errorf("failed to enable forwarding: %w", err)
		}
	}

	if host.Filter, err = NewFilter(host.nftConn); err != nil {
		return nil, fmt.Errorf("failed to setup nftables: %w", err)
	}

	for _, r := range host.FilterRules {
		host.Filter.AddRule(r.Hook, r.Exprs...)
	}

	if err := host.Filter.Flush(); err != nil {
		return nil, fmt.Errorf("failed to configure nftables: %w", err)
	}

	return host, nil
}

// ConfigureLinks adds links to other nodes which
// have been configured by functional options
func (h *Host) ConfigureLinks() error {
	for _, intf := range h.ConfiguredInterfaces {
		peerDev := fmt.Sprintf("veth-%s", h.Name())

		right := &Interface{
			Name: peerDev,
			Node: intf.Node,
		}

		left := intf
		left.Node = h

		if err := h.network.AddLink(left, right); err != nil {
			return err
		}
	}

	return nil
}

func (h *Host) ConfigureInterface(i *Interface) error {
	h.logger.Info("Configuring interface", zap.Any("intf", i))

	// Disable duplicate address detection (DAD) before adding addresses
	// so we dont end up with tentative addresses and slow test executions
	if !i.EnableDAD {
		fn := filepath.Join("/proc/sys/net/ipv6/conf", i.Name, "accept_dad")
		if err := h.WriteProcFS(fn, "0"); err != nil {
			return fmt.Errorf("failed to enabled IPv6 forwarding: %s", err)
		}
	}

	for _, addr := range i.Addresses {
		if err := h.LinkAddAddress(i.Name, addr); err != nil {
			return fmt.Errorf("failed to add link address: %s", err)
		}
	}

	return h.BaseNode.ConfigureInterface(i)
}

func (h *Host) Traceroute(o *Host, opts ...any) error {
	if h.network != o.network {
		return fmt.Errorf("hosts must be on same network")
	}

	opts = append(opts, o)
	_, _, err := h.Run("traceroute", opts...)
	return err
}

func (h *Host) LookupAddress(n string) *net.IPAddr {
	for _, i := range h.Interfaces {
		if i.IsLoopback() {
			continue
		}

		for _, a := range i.Addresses {
			ip := &net.IPAddr{
				IP: a.IP,
			}

			isV4 := len(a.IP.To4()) == net.IPv4len

			switch n {
			case "ip":
				return ip
			case "ip4":
				if isV4 {
					return ip
				}
			case "ip6":
				if !isV4 {
					return ip
				}
			}
		}
	}

	return nil
}
