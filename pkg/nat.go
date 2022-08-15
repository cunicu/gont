package gont

import (
	"fmt"

	nft "github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

type NATOption interface {
	Apply(n *NAT)
}

type NAT struct {
	*Router

	Table       *nft.Table
	Input       *nft.Chain
	Forward     *nft.Chain
	PostRouting *nft.Chain
}

func (n *NAT) Apply(i *Interface) {
	i.Node = n
}

func (n *Network) AddNAT(name string, opts ...Option) (*NAT, error) {
	rtr, err := n.AddRouter(name, opts...)
	if err != nil {
		return nil, err
	}

	nat := &NAT{
		Router: rtr,
	}

	// Apply NAT options
	for _, o := range opts {
		switch opt := o.(type) {
		case NATOption:
			opt.Apply(nat)
		}
	}

	if err := nat.setupTable(nat.nftConn); err != nil {
		return nil, fmt.Errorf("failed to setup nftables table: %w", err)
	}

	n.Register(nat)

	return nat, nil
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
	}

	// Apply NAT and BaseNode options
	for _, o := range opts {
		switch opt := o.(type) {
		case NATOption:
			opt.Apply(nat)
		case BaseNodeOption:
			opt.Apply(host.BaseNode)
		}
	}

	if err := host.ConfigureLinks(); err != nil {
		return nil, err
	}

	if err := nat.setupTable(n.HostNode.nftConn); err != nil {
		return nil, fmt.Errorf("failed to setup nftables table: %w", err)
	}

	n.Register(host)

	return nat, nil
}

/* Setup the table
 *
 * $ nft list table inet gont-nat
 * table inet gont-nat {
 * 	chain input {
 * 		type filter hook input priority filter; policy drop;
 * 	}
 *
 * 	chain forward {
 * 		type filter hook forward priority filter; policy drop;
 * 		iifgroup "south-bound" accept
 * 		ct state established,related accept
 * 	}
 *
 * 	chain snat {
 * 		type nat hook postrouting priority srcnat; policy accept;
 * 		oifgroup "north-bound" masquerade
 * 	}
 * }
 */
func (n *NAT) setupTable(c *nft.Conn) error {
	chainPolicyDrop := nft.ChainPolicyDrop

	// Delete any old table
	t := &nft.Table{
		Family: nft.TableFamilyINet,
		Name:   "gont-nat",
	}

	c.DelTable(t)
	c.Flush()
	// We ignore the error here, as DelTable might fail if there is no old table existing

	n.Table = c.AddTable(t)

	// Input chain
	n.Input = c.AddChain(&nft.Chain{
		Name:     "input",
		Table:    n.Table,
		Type:     nft.ChainTypeFilter,
		Hooknum:  nft.ChainHookInput,
		Priority: nft.ChainPriorityFilter,

		// Drop policy is crucial here as it avoid ICMP port-unreachable
		// messages during UDP hole punching.
		// See: https://www.spinics.net/lists/netfilter/msg58226.html
		Policy: &chainPolicyDrop,
	})

	// icmp6 accept
	c.AddRule(&nft.Rule{
		Table: n.Table,
		Chain: n.Input,
		Exprs: []expr.Any{
			&expr.Meta{
				Key:      expr.MetaKeyL4PROTO,
				Register: 1,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data: binaryutil.NativeEndian.PutUint16(
					unix.IPPROTO_ICMPV6,
				),
			},
			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})

	// icmp accept
	c.AddRule(&nft.Rule{
		Table: n.Table,
		Chain: n.Input,
		Exprs: []expr.Any{
			&expr.Meta{
				Key:      expr.MetaKeyL4PROTO,
				Register: 1,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data: binaryutil.NativeEndian.PutUint16(
					unix.IPPROTO_ICMP,
				),
			},
			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})

	// Forward chain
	n.Forward = c.AddChain(&nft.Chain{
		Name:     "forward",
		Table:    n.Table,
		Type:     nft.ChainTypeFilter,
		Hooknum:  nft.ChainHookForward,
		Priority: nft.ChainPriorityFilter,
		Policy:   &chainPolicyDrop,
	})

	c.AddRule(&nft.Rule{
		Table: n.Table,
		Chain: n.Forward,
		Exprs: []expr.Any{
			&expr.Meta{
				Key:      expr.MetaKeyIIFGROUP,
				Register: 1,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data: binaryutil.NativeEndian.PutUint32(
					uint32(DeviceGroupSouthBound),
				),
			},
			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})

	c.AddRule(&nft.Rule{
		Table: n.Table,
		Chain: n.Forward,
		Exprs: []expr.Any{
			&expr.Ct{
				Register: 1,
				Key:      expr.CtKeySTATE,
			},
			&expr.Bitwise{
				SourceRegister: 1,
				DestRegister:   1,
				Len:            4,
				Mask:           binaryutil.NativeEndian.PutUint32(expr.CtStateBitESTABLISHED | expr.CtStateBitRELATED),
				Xor:            binaryutil.NativeEndian.PutUint32(0),
			},
			&expr.Cmp{
				Op:       expr.CmpOpNeq,
				Register: 1,
				Data:     []byte{0, 0, 0, 0},
			},
			&expr.Verdict{
				Kind: expr.VerdictAccept,
			},
		},
	})

	// Postrouting chain
	n.PostRouting = c.AddChain(&nft.Chain{
		Name:     "snat",
		Table:    n.Table,
		Type:     nft.ChainTypeNAT,
		Hooknum:  nft.ChainHookPostrouting,
		Priority: nft.ChainPriorityNATSource,
	})

	c.AddRule(&nft.Rule{
		Table: n.Table,
		Chain: n.PostRouting,
		Exprs: []expr.Any{
			&expr.Meta{
				Key:      expr.MetaKeyOIFGROUP,
				Register: 1,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data: binaryutil.NativeEndian.PutUint32(
					uint32(DeviceGroupNorthBound),
				),
			},
			&expr.Masq{},
		},
	})

	return c.Flush()
}
