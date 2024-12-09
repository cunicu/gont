// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"

	g "cunicu.li/gont/v2/pkg"
	sdbus "github.com/coreos/go-systemd/v22/dbus"
	"github.com/godbus/dbus/v5"
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

	sliceName := fmt.Sprintf("gont-%s-%s", network, node)
	scopeName := fmt.Sprintf("gont-run-%d", os.Getpid())

	cgroup, err := g.NewCGroup(nil, "scope", scopeName)
	if err != nil {
		return fmt.Errorf("failed to create cgroup: %w", err)
	}

	cgroup.Properties = append(cgroup.Properties,
		sdbus.Property{
			Name:  "Slice",
			Value: dbus.MakeVariant(sliceName + ".slice"),
		},
		sdbus.Property{
			Name:  "PIDs",
			Value: dbus.MakeVariant([]uint{uint(os.Getpid())}),
		},
	)

	if err := cgroup.Start(); err != nil {
		return fmt.Errorf("failed to start cgroup: %w", err)
	}

	return g.Exec(network, node, args)
}
