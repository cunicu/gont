// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Limit debugger support to platforms which are supported by Delve
// See: https://github.com/go-delve/delve/blob/master/pkg/proc/native/support_sentinel_linux.go
//go:build linux && (amd64 || arm64 || 386)

package gont

import (
	"fmt"
	"net"
	"strconv"

	"github.com/go-delve/delve/service/api"
	"go.uber.org/zap"
)

type DebuggerOption interface {
	ApplyDebugger(*Debugger)
}

type Debugger struct {
	// Options
	BreakOnEntry         bool
	DetachOnExit         bool
	Tracepoints          []Tracepoint
	Tracers              []*Tracer
	ListenAddr           string
	DebugInfoDirectories []string

	listenAddr *net.TCPAddr

	instances map[int]*debuggerInstance
	stop      chan struct{}
	logger    *zap.Logger
}

func (d *Debugger) ApplyNetwork(n *Network) {
	n.Debugger = d
}

func (d *Debugger) ApplyBaseNode(n *BaseNode) {
	n.Debugger = d
}

func (d *Debugger) ApplyCmd(c *Cmd) {
	c.Debugger = d
}

func NewDebugger(opts ...DebuggerOption) *Debugger {
	d := &Debugger{
		DebugInfoDirectories: []string{"/usr/lib/debug/.build-id"},
		DetachOnExit:         true,

		instances: map[int]*debuggerInstance{},
		stop:      make(chan struct{}),
		logger:    zap.L().Named("debug"),
	}

	for _, opt := range opts {
		opt.ApplyDebugger(d)
	}

	if d.ListenAddr != "" {
		host, port, err := net.SplitHostPort(d.ListenAddr)
		if err != nil {
			panic(err)
		}

		portnum, err := strconv.Atoi(port)
		if err != nil {
			panic(err)
		}

		d.listenAddr = &net.TCPAddr{
			IP:   net.ParseIP(host),
			Port: portnum,
		}
	}

	return d
}

func (d *Debugger) Close() error {
	close(d.stop)

	for _, dbg := range d.instances {
		if dbg.IsRunning() {
			if _, err := dbg.Command(&api.DebuggerCommand{Name: api.Halt}, nil, nil); err != nil {
				return fmt.Errorf("failed to halt process: %w", err)
			}
		}

		if err := dbg.Detach(false); err != nil {
			return fmt.Errorf("failed to detach from process: %w", err)
		}
	}

	return nil
}

func (d *Debugger) nextListenAddr() *net.TCPAddr {
	if d.listenAddr == nil {
		return nil
	}

	return &net.TCPAddr{
		IP:   d.listenAddr.IP,
		Zone: d.listenAddr.Zone,
		Port: d.listenAddr.Port + len(d.instances),
	}
}
