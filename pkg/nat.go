package gont

import (
	"bytes"
	"fmt"
)

const (
	sbSet = "sb"
)

type NAT struct {
	Router
}

func (n *Network) AddNAT(name string, opts ...Option) (*NAT, error) {
	rtr, err := n.AddRouter(name, opts...)
	if err != nil {
		return nil, err
	}

	nat := &NAT{
		Router: *rtr,
	}

	// Apply NAT options
	for _, opt := range opts {
		if nopt, ok := opt.(NATOption); ok {
			nopt.Apply(nat)
		}
	}

	return nat, nat.setup()
}

func (n *Network) AddHostNAT(name string, opts ...Option) (*NAT, error) {
	host := n.HostNode

	if err := host.EnableForwarding(); err != nil {
		return nil, err
	}

	rtr := &Router{
		Host: *host,
	}

	nat := &NAT{
		Router: *rtr,
	}

	// Apply NAT options
	for _, opt := range opts {
		if nopt, ok := opt.(NATOption); ok {
			nopt.Apply(nat)
		}
	}

	if err := nat.setup(); err != nil {
		return nil, err
	}

	// Dummy host for getting interfaces
	h := &Host{}
	for _, opt := range opts {
		if nopt, ok := opt.(HostOption); ok {
			nopt.Apply(h)
		}
	}

	for _, i := range h.Interfaces {
		nat.AddInterface(i)
	}

	return nat, nil
}

func (n *NAT) setup() error {
	var err error

	sbGroup := fmt.Sprintf("%d", NATSouthBound)

	// Setup ipset of all south-bound networks
	// TODO: use netlink interface for configuring ipsets
	if out, _, err := n.Run("ipset", "create", sbSet, "hash:net"); err != nil {
		if !bytes.Contains(out, []byte("already exists")) {
			return err
		}
	}

	for _, i := range n.Interfaces {
		n.updateIPSetInterface(i)
	}

	// Setup NAT rules in iptables
	if _, _, err = n.Run("iptables", "--insert", "FORWARD", "-m", "devgroup", "--src-group", sbGroup, "--match", "set", "--match-set", sbSet, "dst", "--jump", "DROP"); err != nil {
		return err
	}
	if _, _, err = n.Run("iptables", "--append", "FORWARD", "-m", "devgroup", "--src-group", sbGroup, "--match", "set", "--match-set", sbSet, "src", "--jump", "ACCEPT"); err != nil {
		return err
	}
	if _, _, err = n.Run("iptables", "--append", "FORWARD", "-m", "devgroup", "--dst-group", sbGroup, "--match", "set", "--match-set", sbSet, "dst", "--jump", "ACCEPT"); err != nil {
		return err
	}
	if _, _, err = n.Run("iptables", "--table", "nat", "--append", "POSTROUTING", "--match", "set", "--match-set", sbSet, "src", "--match", "set", "!", "--match-set", sbSet, "dst", "--jump", "MASQUERADE"); err != nil {
		return err
	}

	return nil
}

func (n *NAT) updateIPSetInterface(i Interface) error {
	if i.Group == NATSouthBound {
		for _, a := range i.Addresses {

			// TODO: use netlink interface for configuring ipsets
			if out, _, err := n.Run("ipset", "add", sbSet, a.String()); err != nil {
				if !bytes.Contains(out, []byte("already exists")) {
					return err
				}
			}
		}
	}

	return nil
}

func (n *NAT) AddInterface(i Interface) error {
	if err := n.Host.AddInterface(i); err != nil {
		return err
	}

	return n.updateIPSetInterface(i)
}
