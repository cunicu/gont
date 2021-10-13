package gont

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	nl "github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
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

	n.Nodes[name] = nat // TODO: quirk to get n.UpdateHostsFile() working

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
	for _, family := range []uint8{unix.AF_INET, unix.AF_INET6} {
		sbSetName := sbSet
		if family == unix.AF_INET {
			sbSetName += "-inet"
		} else {
			sbSetName += "-inet6"
		}

		log.WithField("set", sbSetName).Info("Creating ipset")
		if err := n.RunFunc(func() error {
			return nl.IpsetCreate(sbSetName, "hash:net", nl.IpsetCreateOptions{
				Replace: true,
				Family:  family,
			})
		}); err != nil {
			return err
		}
	}

	for _, i := range n.Interfaces {
		n.updateIPSetInterface(i)
	}

	// Setup NAT rules in iptables

	for _, family := range []string{"inet", "inet6"} {
		sbSetName := fmt.Sprintf("%s-%s", sbSet, family)
		ipt := "iptables"
		if family == "inet6" {
			ipt = "ip6tables"
		}

		if _, _, err = n.Run(ipt, "--insert", "FORWARD", "-m", "devgroup", "--src-group", sbGroup, "--match", "set", "--match-set", sbSetName, "dst", "--jump", "DROP"); err != nil {
			return err
		}
		if _, _, err = n.Run(ipt, "--append", "FORWARD", "-m", "devgroup", "--src-group", sbGroup, "--match", "set", "--match-set", sbSetName, "src", "--jump", "ACCEPT"); err != nil {
			return err
		}
		if _, _, err = n.Run(ipt, "--append", "FORWARD", "-m", "devgroup", "--dst-group", sbGroup, "--match", "set", "--match-set", sbSetName, "dst", "--jump", "ACCEPT"); err != nil {
			return err
		}
		if _, _, err = n.Run(ipt, "--table", "nat", "--append", "POSTROUTING", "--match", "set", "--match-set", sbSetName, "src", "--match", "set", "!", "--match-set", sbSetName, "dst", "--jump", "MASQUERADE"); err != nil {
			return err
		}
	}

	return nil
}

func (n *NAT) updateIPSetInterface(i Interface) error {
	if i.Group == NATSouthBound {
		for _, a := range i.Addresses {
			family := unix.AF_INET
			if a.IP.To4() == nil {
				family = unix.AF_INET6
			}

			sbSetName := sbSet
			if family == unix.AF_INET {
				sbSetName += "-inet"
			} else {
				sbSetName += "-inet6"
			}

			log.WithFields(log.Fields{
				"set":  sbSetName,
				"addr": a.String(),
			}).Info("Adding address to ipset")

			cidr, _ := a.Mask.Size()
			if err := n.RunFunc(func() error {
				return nl.IpsetAdd(sbSetName, &nl.IPSetEntry{
					IP:      a.IP.Mask(a.Mask),
					CIDR:    uint8(cidr),
					Comment: fmt.Sprintf("gont:%s/%s", n.name, i.Name),
				})
			}); err != nil {
				return err
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
