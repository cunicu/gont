// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"cunicu.li/gont/v2/internal/utils"
)

func version() {
	version := "unknown"
	if tag != "" {
		version = tag
	}

	if ok, rev, dirty, btime := utils.ReadVCSInfos(); ok {
		dirtyFlag := ""
		if dirty {
			dirtyFlag = "-dirty"
		}

		fmt.Printf("%s (%s%s, build on %s)\n", version, rev[:8], dirtyFlag, btime.String())
	} else {
		fmt.Println(version)
	}
}
