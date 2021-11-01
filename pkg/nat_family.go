package gont

import (
	"net"

	nft "github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
)

type natFamily struct {
	Family      nft.TableFamily
	SetDataType nft.SetDatatype

	Table       *nft.Table
	Forward     *nft.Chain
	PostRouting *nft.Chain
	Set         *nft.Set
}

func newNATFamily(f nft.TableFamily) *natFamily {
	g := &natFamily{
		Family: f,
	}

	switch f {
	case nft.TableFamilyIPv4:
		g.SetDataType = nft.TypeIPAddr
	case nft.TableFamilyIPv6:
		g.SetDataType = nft.TypeIP6Addr
	}

	return g
}

func (f *natFamily) SetupTable(c *nft.Conn) error {
	t := &nft.Table{
		Family: f.Family,
		Name:   "gont",
	}

	c.DelTable(t)
	c.Flush()

	f.Table = c.AddTable(t)

	return c.Flush()
}

func (f *natFamily) SetupSet(c *nft.Conn) error {
	f.Set = &nft.Set{
		Name:     "sb",
		Table:    f.Table,
		KeyType:  f.SetDataType,
		Interval: true,
	}

	if err := c.AddSet(f.Set, []nft.SetElement{}); err != nil {
		return err
	}

	return c.Flush()
}

func (f *natFamily) SetupChains(c *nft.Conn, sbGroup uint32) error {
	var saddr, daddr *expr.Payload
	switch f.Family {
	case nft.TableFamilyIPv4:
		saddr = &expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       12,
			Len:          net.IPv4len,
		}

		daddr = &expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       16,
			Len:          net.IPv4len,
		}
	case nft.TableFamilyIPv6:
		saddr = &expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       8,
			Len:          net.IPv6len,
		}

		daddr = &expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       8 + 16,
			Len:          net.IPv6len,
		}
	}

	accept := &expr.Verdict{Kind: expr.VerdictAccept}
	drop := &expr.Verdict{Kind: expr.VerdictDrop}

	is_southbound_dev := &expr.Cmp{
		Op:       expr.CmpOpEq,
		Register: 1,
		Data:     binaryutil.NativeEndian.PutUint32(sbGroup),
	}

	is_southbound_net := &expr.Lookup{
		SourceRegister: 1,
		SetName:        f.Set.Name,
	}

	is_not_southbound_net := &expr.Lookup{
		SourceRegister: 1,
		SetName:        f.Set.Name,
		Invert:         true,
	}

	iifgroup := &expr.Meta{
		Key:      expr.MetaKeyIIFGROUP,
		Register: 1,
	}

	oifgroup := &expr.Meta{
		Key:      expr.MetaKeyOIFGROUP,
		Register: 1,
	}

	// Forward chain
	f.Forward = c.AddChain(&nft.Chain{
		Name:     "forward",
		Table:    f.Table,
		Type:     nft.ChainTypeFilter,
		Hooknum:  nft.ChainHookForward,
		Priority: nft.ChainPriorityFilter,
	})

	c.AddRule(&nft.Rule{
		Table: f.Table,
		Chain: f.Forward,
		Exprs: []expr.Any{
			iifgroup, is_southbound_dev, // meta iifgroup <sbGroup>
			daddr, is_southbound_net, // ip daddr @sb
			drop,
		},
	})

	c.AddRule(&nft.Rule{
		Table: f.Table,
		Chain: f.Forward,
		Exprs: []expr.Any{
			iifgroup, is_southbound_dev, // meta iifgroup <sbGroup>
			saddr, is_southbound_net, // ip saddr @sb
			accept,
		},
	})

	c.AddRule(&nft.Rule{
		Table: f.Table,
		Chain: f.Forward,
		Exprs: []expr.Any{
			oifgroup, is_southbound_dev, // meta oifgroup <sbGroup>
			daddr, is_not_southbound_net, // ip daddr @sb
			accept,
		},
	})

	// Postrouting chain
	postrouting := c.AddChain(&nft.Chain{
		Name:     "snat",
		Table:    f.Table,
		Type:     nft.ChainTypeNAT,
		Hooknum:  nft.ChainHookPostrouting,
		Priority: nft.ChainPriorityNATSource,
	})

	c.AddRule(&nft.Rule{
		Table: f.Table,
		Chain: postrouting,
		Exprs: []expr.Any{
			saddr, is_southbound_net, // ip saddr @sb
			daddr, is_not_southbound_net, // ip daddr != @sb
			&expr.Masq{}, // masquerade
		},
	})

	return c.Flush()
}

func (f *natFamily) modifyNetwork(c *nft.Conn, netw net.IPNet, op func(s *nft.Set, vals []nft.SetElement) error) error {

	start, end := ipNetNextRange(netw)

	startElm := nft.SetElement{
		Key:         f.ipBytes(start),
		IntervalEnd: false,
	}

	endElm := nft.SetElement{
		Key:         f.ipBytes(end),
		IntervalEnd: true,
	}

	return op(f.Set, []nft.SetElement{startElm, endElm})
}

func (f *natFamily) ipBytes(i net.IP) []byte {
	switch f.Family {
	case nft.TableFamilyIPv4:
		return i.To4()

	case nft.TableFamilyIPv6:
		return i.To16()

	default:
		return []byte{}
	}
}

func (f *natFamily) AddNetwork(c *nft.Conn, netw net.IPNet) error {
	return f.modifyNetwork(c, netw, c.SetAddElements)
}

func (f *natFamily) DeleteNetwork(c *nft.Conn, netw net.IPNet) error {
	return f.modifyNetwork(c, netw, c.SetDeleteElements)
}
