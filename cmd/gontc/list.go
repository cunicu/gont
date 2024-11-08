// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	g "cunicu.li/gont/v2/pkg"
)

func list(args []string) {
	if len(args) > 1 {
		network := args[1]
		for _, name := range g.NodeNames(network) {
			fmt.Printf("%s/%s\n", network, name)
		}
	} else {
		for _, name := range g.NetworkNames() {
			fmt.Println(name)
		}
	}
}
