// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package filters contains the options for configuring NFTables filtering
package filters

import "github.com/google/nftables/expr"

// Statement is a list of one or more nftables expressions
type Statement []expr.Any

// Drop is a statement which drops all packets
var Drop = Statement{ //nolint:gochecknoglobals
	&expr.Verdict{
		Kind: expr.VerdictDrop,
	},
}
