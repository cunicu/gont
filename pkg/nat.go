package gont

import (
	"fmt"
	"net"

	nft "github.com/google/nftables"
	log "github.com/sirupsen/logrus"
)

var (
	families = []nft.TableFamily{
		nft.TableFamilyIPv4,
		nft.TableFamilyIPv6,
	}
)

type NAT struct {
	*Router

	families map[nft.TableFamily]*natFamily
}

func (h *NAT) Apply(i *Interface) {
	i.Node = h
}

func (n *Network) AddNAT(name string, opts ...Option) (*NAT, error) {
	rtr, err := n.AddRouter(name, opts...)
	if err != nil {
		return nil, err
	}

	nat := &NAT{
		Router: rtr,

		families: map[nft.TableFamily]*natFamily{},
	}

	n.Register(nat)

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
		Host: host,
	}

	nat := &NAT{
		Router: rtr,

		families: map[nft.TableFamily]*natFamily{},
	}

	// Apply NAT options
	for _, o := range opts {
		switch opt := o.(type) {
		case NATOption:
			opt.Apply(nat)
		case NodeOption:
			opt.Apply(host.BaseNode)
		}
	}

	if err := host.ConfigureLinks(); err != nil {
		return nil, err
	}

	if err := nat.setup(); err != nil {
		return nil, err
	}

	n.Register(host)

	return nat, nil
}

func (n *NAT) setup() error {
	var sbGroup uint32 = uint32(DeviceGroupSouthBound)

	c := &nft.Conn{
		NetNS: int(n.NsHandle),
	}

	for _, g := range families {
		f := newNATFamily(g)

		if err := f.SetupTable(c); err != nil {
			return err
		}
		if err := f.SetupSet(c); err != nil {
			return err
		}
		if err := f.SetupChains(c, sbGroup); err != nil {
			return err
		}

		n.families[g] = f
	}

	for _, i := range n.Interfaces {
		if err := n.updateIPSetInterface(c, i); err != nil {
			return err
		}
	}

	if err := c.Flush(); err != nil {
		return fmt.Errorf("failed setup nftables: %w", err)
	}

	return nil
}

func (n *NAT) updateIPSetInterface(c *nft.Conn, i *Interface) error {
	if i.LinkAttrs.Group == uint32(DeviceGroupSouthBound) {
		for _, a := range i.Addresses {
			f := n.families[nft.TableFamilyIPv4]
			if a.IP.To4() == nil {
				f = n.families[nft.TableFamilyIPv6]
			}

			log.WithFields(log.Fields{
				"set":  f.Set.Name,
				"addr": a.String(),
			}).Info("Adding address to nftables set")

			netw := net.IPNet{
				IP:   a.IP.Mask(a.Mask),
				Mask: a.Mask,
			}

			if err := f.AddNetwork(c, netw); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *NAT) ConfigureInterface(i *Interface) error {
	c := &nft.Conn{
		NetNS: int(n.NsHandle),
	}

	if err := n.updateIPSetInterface(c, i); err != nil {
		return err
	}

	if err := c.Flush(); err != nil {
		return err
	}

	return n.Router.ConfigureInterface(i)
}

func ipNetNextRange(netw net.IPNet) (net.IP, net.IP) {
	start := netw.IP
	ones, _ := netw.Mask.Size()

	bp := ones / 8
	bm := ones % 8

	if bm == 0 {
		bp -= 1
		bm = 8
	}

	end := make(net.IP, len(start))
	copy(end, start)

	end[bp] = end[bp] + 1<<(8-bm)

	return start, end
}
