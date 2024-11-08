// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"

	g "cunicu.li/gont/v2/pkg"
)

func exec(network, node string, args []string) error {
	if len(flag.Args()) <= 1 {
		return fmt.Errorf("not enough arguments")
	}

	if network == "" {
		return fmt.Errorf("there is no active Gont network")
	}

	if err := os.Setenv("GONT_NETWORK", network); err != nil {
		return err
	}
	if err := os.Setenv("GONT_NODE", node); err != nil {
		return err
	}

	cgroupName := fmt.Sprintf("gont-run-%d", os.Getpid())
	cgroup, err := g.NewCGroup(nil, "scope", cgroupName)
	if err != nil {
		return fmt.Errorf("failed to create cgroup: %w", err)
	}

	if err := cgroup.Start(); err != nil {
		return fmt.Errorf("failed to start cgroup: %w", err)
	}

	return g.Exec(network, node, args)
}
