// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"time"

	g "cunicu.li/gont/v2/pkg"
	"github.com/coreos/go-systemd/v22/dbus"
)

func clean(args []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second) //nolint:govet
	defer cancel()

	c, err := dbus.NewWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to D-Bus: %w", err)
	}

	networks := args[1:]
	if len(networks) == 0 {
		networks = g.NetworkNames()
	}

	for _, name := range networks {
		if err := g.TeardownNetwork(ctx, c, name); err != nil {
			return fmt.Errorf("failed to teardown network '%s': %w", name, err)
		}

		fmt.Println(name)
	}

	return nil
}
