// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	nft "github.com/google/nftables"
	"github.com/google/nftables/expr"
)

type FilterHook int

const (
	FilterInput FilterHook = iota
	FilterOutput
	FilterForward
)

type FilterRule struct {
	Exprs []expr.Any

	Hook FilterHook
}

func (fr FilterRule) Apply(h *Host) {
	h.FilterRules = append(h.FilterRules, &fr)
}

type Filter struct {
	conn *nft.Conn

	Family nft.TableFamily
	Table  *nft.Table

	Input   *nft.Chain
	Output  *nft.Chain
	Forward *nft.Chain
}

func NewFilter(c *nft.Conn) (*Filter, error) {
	flt := &Filter{
		conn: c,

		Family: nft.TableFamilyINet,
	}

	t := &nft.Table{
		Family: flt.Family,
		Name:   "gont",
	}

	c.DelTable(t)
	c.Flush()
	// We ignore the error here, as DelTable might fail if there is no old table existing

	flt.Table = c.AddTable(t)

	if err := c.Flush(); err != nil {
		return nil, err
	}

	flt.Input = c.AddChain(&nft.Chain{
		Name:     "input",
		Table:    flt.Table,
		Type:     nft.ChainTypeFilter,
		Hooknum:  nft.ChainHookInput,
		Priority: nft.ChainPriorityFilter,
	})

	flt.Output = c.AddChain(&nft.Chain{
		Name:     "output",
		Table:    flt.Table,
		Type:     nft.ChainTypeFilter,
		Hooknum:  nft.ChainHookOutput,
		Priority: nft.ChainPriorityFilter,
	})

	flt.Forward = c.AddChain(&nft.Chain{
		Name:     "forward",
		Table:    flt.Table,
		Type:     nft.ChainTypeFilter,
		Hooknum:  nft.ChainHookForward,
		Priority: nft.ChainPriorityFilter,
	})

	return flt, c.Flush()
}

func (f *Filter) AddRule(h FilterHook, exprs ...expr.Any) {
	var chain *nft.Chain
	switch h {
	case FilterForward:
		chain = f.Forward
	case FilterInput:
		chain = f.Input
	case FilterOutput:
		chain = f.Output
	}

	f.conn.AddRule(&nft.Rule{
		Table: f.Table,
		Chain: chain,
		Exprs: exprs,
	})
}

func (f *Filter) Flush() error {
	return f.conn.Flush()
}
