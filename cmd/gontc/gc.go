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

func collectGarbage(_ []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second) //nolint:govet
	defer cancel()

	c, err := dbus.NewWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to D-Bus: %w", err)
	}

	deleted, err := g.TeardownStaleCgroups(ctx, c)
	if err != nil {
		return err
	}

	for _, name := range deleted {
		fmt.Println(name)
	}

	return nil
}
