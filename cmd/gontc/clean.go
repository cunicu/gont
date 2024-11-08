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
	ctx, _ = context.WithTimeout(ctx, 10*time.Second)

	c, err := dbus.NewWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to D-Bus: %w", err)
	}

	if len(args) > 1 {
		network := args[1]
		if err := g.TeardownNetwork(ctx, c, network); err != nil {
			return fmt.Errorf("failed to teardown network '%s': %w", network, err)
		}
	} else {
		return g.TeardownAllNetworks(ctx, c)
	}

	return nil
}
