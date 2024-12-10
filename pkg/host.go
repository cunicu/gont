// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"
	"net"
	"path/filepath"

	nl "github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

type HostOption interface {
	ApplyHost(h *Host)
}

type Host struct {
	*BaseNode

	Filter *Filter

	// Options
	FilterRules []*FilterRule
	Routes      []*nl.Route
}

// Options
func (h *Host) ApplyInterface(i *Interface) {
	i.Node = h
}

func (n *Network) AddHost(name string, opts ...Option) (h *Host, err error) {
	h = &Host{}

	if h.BaseNode, err = n.AddNode(name, opts...); err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	n.Register(h)

	// Apply host options
	for _, opt := range opts {
		if hOpt, ok := opt.(HostOption); ok {
			hOpt.ApplyHost(h)
		}
	}

	// Configure loopback device
	if !h.Namespace.IsHost() {
		lo := Interface{
			Name: loopbackInterfaceName,
			Node: h,
			Addresses: []net.IPNet{
				{
					IP:   net.IPv4(127, 0, 0, 1),
					Mask: net.IPv4Mask(255, 0, 0, 0),
				},
				{
					IP:   net.IPv6loopback,
					Mask: net.CIDRMask(8*net.IPv6len, 8*net.IPv6len),
				},
			},
		}

		if lo.Link, err = h.nlHandle.LinkByName("lo"); err != nil {
			return nil, fmt.Errorf("failed to get loopback interface: %w", err)
		}

		if err := h.ConfigureInterface(&lo); err != nil {
			return nil, fmt.Errorf("failed to configure loopback interface: %w", err)
		}
	}

	if err := h.ConfigureLinks(); err != nil {
		return nil, fmt.Errorf("failed to configure links: %w", err)
	}

	// Configure host
	for _, r := range h.Routes {
		if err := h.AddRoute(r); err != nil {
			return nil, fmt.Errorf("failed to add route: %w", err)
		}
	}

	// Setup nftables filters
	if h.Filter, err = NewFilter(h.nftConn); err != nil {
		return nil, fmt.Errorf("failed to setup nftables: %w", err)
	}

	for _, r := range h.FilterRules {
		h.Filter.AddRule(r.Hook, r.Exprs...)
	}

	if err := h.Filter.Flush(); err != nil {
		return nil, fmt.Errorf("failed to configure nftables: %w", err)
	}

	return h, nil
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
	// so we do not end up with tentative addresses and slow test executions
	if !i.EnableDAD {
		fn := filepath.Join("/proc/sys/net/ipv6/conf", i.Name, "accept_dad")
		if err := h.WriteProcFS(fn, "0"); err != nil {
			return fmt.Errorf("failed to enabled IPv6 forwarding: %w", err)
		}
	}

	for _, addr := range i.Addresses {
		if err := i.AddAddress(&addr); err != nil {
			return fmt.Errorf("failed to add link address: %w", err)
		}
	}

	return h.BaseNode.ConfigureInterface(i)
}

func (h *Host) Traceroute(o *Host, opts ...any) error {
	if h.network != o.network {
		return errInvalidNetwork
	}

	opts = append(opts, o)
	_, err := h.Run("traceroute", opts...)
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
