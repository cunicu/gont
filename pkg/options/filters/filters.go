// Package filters contains the options for configuring NFTables filtering
package filters

import "github.com/google/nftables/expr"

// Statement is a list of one or more nftables expressions
type Statement []expr.Any

var Drop = Statement{
	&expr.Verdict{
		Kind: expr.VerdictDrop,
	},
}
