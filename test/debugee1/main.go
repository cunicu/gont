// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"runtime"
)

var myID string //nolint:unused,gochecknoglobals

func main() {
	myID = os.Args[1]

	runtime.Breakpoint()
}
