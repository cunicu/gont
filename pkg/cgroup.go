// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
)

var errWaitCGroupShutdown = errors.New("failed to wait for CGroup shutdown")

type CGroupOption interface {
	ApplyCGroup(s *CGroup)
}

type CGroup struct {
	Name       string
	Type       string
	Properties []dbus.Property

	sdConn *dbus.Conn
}

func NewCGroup(c *dbus.Conn, typ, name string, opts ...Option) (g *CGroup, err error) {
	g = &CGroup{
		Name:   name,
		Type:   typ,
		sdConn: c,
	}

	// Create D-Bus connection if not existing
	if g.sdConn == nil {
		if g.sdConn, err = dbus.NewWithContext(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to connect to D-Bus: %w", err)
		}
	}

	for _, opt := range opts {
		if opt, ok := opt.(CGroupOption); ok {
			opt.ApplyCGroup(g)
		}
	}

	return g, nil
}

func (g *CGroup) Unit() string {
	return g.Name + "." + g.Type
}

// Start creates the CGroup
func (g *CGroup) Start() error {
	ch := make(chan string)
	if _, err := g.sdConn.StartTransientUnitContext(context.Background(), g.Unit(), "replace", g.Properties, ch); err != nil {
		return fmt.Errorf("failed to create slice: %w", err)
	}

	<-ch

	return nil
}

// Stop stops the CGroup and kills all contained processes
func (g *CGroup) Stop() error {
	ch := make(chan string)
	if _, err := g.sdConn.StopUnitContext(context.Background(), g.Unit(), "fail", ch); err != nil {
		return fmt.Errorf("failed to remove slice: %w", err)
	}

	if state := <-ch; state != "done" {
		return fmt.Errorf("%w: state is %s", errWaitCGroupShutdown, state)
	}

	return nil
}

// Freeze suspends execution of all processes in the control group.
func (g *CGroup) Freeze() error {
	return g.sdConn.FreezeUnit(context.Background(), g.Unit())
}

// Thaw resumes execution of all processes in the control group.
func (g *CGroup) Thaw() error {
	return g.sdConn.ThawUnit(context.Background(), g.Unit())
}

// SetProperties sets transient systemd CGroup properties of the unit.
// See: https://systemd.io/TRANSIENT-SETTINGS/
func (g *CGroup) SetProperties(opts ...CGroupOption) error {
	so := &CGroup{}
	for _, opt := range opts {
		opt.ApplyCGroup(g)
		opt.ApplyCGroup(so)
	}

	return g.sdConn.SetUnitPropertiesContext(context.Background(), g.Unit(), true, so.Properties...)
}
