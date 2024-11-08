// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
)

func shell(network, node string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	ps1 := fmt.Sprintf("%s/%s: ", network, node)
	os.Setenv("PS1", ps1)

	cmd := []string{shell, "--norc"}

	return exec(network, node, cmd)
}
