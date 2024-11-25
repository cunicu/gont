// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Limit debugger support to platforms which are supported by Delve
// See: https://github.com/go-delve/delve/blob/master/pkg/proc/native/support_sentinel_linux.go
//go:build linux && (amd64 || arm64 || 386)

package gont

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/service"
	"github.com/go-delve/delve/service/api"
	"github.com/go-delve/delve/service/dap"
	debug "github.com/go-delve/delve/service/debugger"
	"go.uber.org/zap"
)

type debuggerInstance struct {
	*debug.Debugger

	debugger   *Debugger
	cmd        *exec.Cmd
	stop       chan any
	listenAddr *net.TCPAddr
	config     service.Config

	control sync.Mutex

	logger *zap.Logger
}

func (d *Debugger) start(c *exec.Cmd) (*debuggerInstance, error) {
	var err error

	di := &debuggerInstance{
		cmd:      c,
		stop:     make(chan any),
		debugger: d,
		config: service.Config{
			Debugger: debug.Config{
				Backend:              "native",
				ExecuteKind:          debug.ExecutingExistingFile,
				DebugInfoDirectories: d.DebugInfoDirectories,
				AttachPid:            c.Process.Pid,
				WorkingDir:           c.Dir,
			},
			ProcessArgs: []string{c.Path},
			AcceptMulti: true,
		},
		listenAddr: d.nextListenAddr(),
		logger: d.logger.With(
			zap.Int("pid", c.Process.Pid),
		),
	}

	di.config.ProcessArgs = append(di.config.ProcessArgs, c.Args...)

	if di.config.Debugger.WorkingDir == "" {
		wd, _ := os.Getwd()
		di.config.Debugger.WorkingDir = wd
	}

	// if err := logflags.Setup(true, "", ""); err != nil {
	// 	return nil, err
	// }

	// Open listeners
	if di.listenAddr != nil {
		go di.listen(di.listenAddr)
	}

	if di.Debugger, err = debug.New(&di.config.Debugger, di.config.ProcessArgs); err != nil {
		return nil, fmt.Errorf("failed to create debugger: %w", err)
	}

	for _, tp := range d.Tracepoints {
		if err := di.createBreakpoints(&tp); err != nil {
			return nil, fmt.Errorf("failed to create breakpoints: %w", err)
		}
	}

	for _, t := range d.Tracers {
		if err := t.start(); err != nil {
			return nil, fmt.Errorf("failed to start tracer: %w", err)
		}
	}

	d.instances[c.Process.Pid] = di

	if d.DetachOnExit {
		if _, err = di.CreateBreakpoint(&api.Breakpoint{
			FunctionName: "runtime.exit",
		}, "", nil, false); err != nil {
			return nil, err
		}
	}

	if d.BreakOnEntry {
		di.logger.Info("Setting break on entry break point")
		if _, err := di.CreateBreakpoint(&api.Breakpoint{
			FunctionName: "runtime.main",
		}, "", nil, false); err != nil {
			return nil, err
		}
	}

	go di.run()

	return di, err
}

func (d *debuggerInstance) run() {
	for {
		d.control.Lock()
		s, err := d.Command(&api.DebuggerCommand{Name: api.Continue}, nil)
		d.control.Unlock()
		if err != nil {
			d.logger.Error("Failed to continue", zap.Error(err))
			break
		}

		reason := d.StopReason()
		d.logger.Debug("Stopped",
			zap.Any("reason", reason),
			zap.Bool("running", s.Running),
			zap.Bool("exited", s.Exited),
		)

		switch reason {
		case proc.StopBreakpoint, proc.StopWatchpoint:
			if d.handleWatchpoints(s) {
				break
			}

			if d.handleBreakpoint(s) {
				return
			}

		case proc.StopUnknown:
			if !s.Exited {
				continue
			}
			fallthrough

		case proc.StopExited:
			d.logger.Debug("Process exited")

			close(d.stop)
			return

		case proc.StopManual:
			d.logger.Debug("Process stopped manually")
			return

		default:
		}
	}
}

func (d *debuggerInstance) handleWatchpoints(s *api.DebuggerState) bool {
	// Re-enable breakpoints for the delayed creation of watchpoints
	for _, wp := range s.WatchOutOfScope {
		if wpi, ok := wp.UserData.(*watchpointInstance); !ok {
			d.logger.Warn("Failed to reenable watchpoint", zap.Any("wp", wp))
			continue
		} else if err := wpi.wentOutOfScope(d); err != nil {
			d.logger.Warn("Failed to enable breakpoint for delayed watchpoint creation", zap.Error(err))
		}
	}

	thr := s.CurrentThread
	if thr.Breakpoint == nil {
		return false
	}

	// Delayed creation of watchpoints
	if bpi, ok := thr.Breakpoint.UserData.(*breakpointInstance); ok && bpi.tracepoint.IsWatchpoint() {
		if err := bpi.createWatchpoint(d, thr); err != nil {
			d.logger.Error("Failed to create watchpoint", zap.Error(err))
		}

		bp := bpi.Breakpoint
		bp.Disabled = true
		if err := d.AmendBreakpoint(bp); err != nil {
			d.logger.Error("Failed to delete breakpoint", zap.Error(err))
		}

		return true
	}

	return false
}

func (d *debuggerInstance) handleBreakpoint(s *api.DebuggerState) bool {
	thr := s.CurrentThread

	if thr.Breakpoint == nil || thr.BreakpointInfo == nil {
		return false
	}

	// Emit trace events for break- and watchpoints
	if tep, ok := thr.Breakpoint.UserData.(traceEventPoint); ok {
		te := tep.traceEvent(thr, d)
		for _, t := range d.debugger.Tracers {
			t.newEvent(te)
		}
	}

	// We detach the debugger here so the final Cmd.Wait() call will
	// populate Cmd.ProcessState properly
	switch thr.Breakpoint.FunctionName {
	case "runtime.exit":
		if d.debugger.DetachOnExit {
			if err := d.Detach(false); err != nil {
				d.logger.Error("Failed to detach", zap.Error(err))
			}

			close(d.stop)
			return true
		}

	case "runtime.main":
		if d.debugger.BreakOnEntry {
			d.logger.Info("Breaking on entry. Stop tracing and waiting for debugger to attach...")
			return true
		}
	}

	return false
}

func (d *debuggerInstance) listen(addr *net.TCPAddr) {
	var err error
	if d.config.Listener, err = net.ListenTCP("tcp", addr); err != nil {
		panic(err)
	}
	defer d.config.Listener.Close()

	for {
		conn, err := d.config.Listener.Accept()
		if err != nil {
			select {
			case <-d.stop:
				// We were supposed to exit, do nothing and return
				return
			default:
				d.logger.Fatal("Failed to listen", zap.Error(err))
			}
		}

		if err := d.haltIfRunning(); err != nil {
			conn.Close()
			d.logger.Error("Failed to take control over process", zap.Error(err))
		}

		s := dap.NewSession(conn, &dap.Config{
			Config:        &d.config,
			StopTriggered: make(chan struct{}),
		}, d.Debugger)

		d.control.Lock()
		d.logger.Debug("Debug session is now controlled by DAP client")

		s.ServeDAPCodec()

		d.control.Unlock()
		d.logger.Debug("Debug session is now controlled by Gont tracer")

		go d.run()
	}
}

func (d *debuggerInstance) haltIfRunning() error {
	if s, err := d.State(true); err != nil {
		return fmt.Errorf("failed to get debugger state: %w", err)
	} else if s.Running {
		if _, err := d.Command(&api.DebuggerCommand{Name: api.Halt}, nil); err != nil {
			return fmt.Errorf("failed to halt process: %w", err)
		}
	}

	return nil
}

func (d *debuggerInstance) createBreakpoints(tp *Tracepoint) error {
	if tp.Location != "" {
		return d.createBreakpointsForLocation(tp)
	}

	return d.createBreakpoint(tp)
}

func (d *debuggerInstance) createBreakpoint(tp *Tracepoint) error {
	var err error

	bpi := &breakpointInstance{
		tracepoint: tp,
	}

	if err := bpi.prepare(0); err != nil {
		return err
	}

	bp := tp.Breakpoint
	bp.UserData = bpi

	if bpi.Breakpoint, err = d.CreateBreakpoint(&bp, "", nil, false); err != nil {
		return fmt.Errorf("failed to create breakpoint: %w", err)
	}

	d.logger.Debug("Created new breakpoint", breakpointFields(bpi.Breakpoint)...)

	return nil
}

func (d *debuggerInstance) createBreakpointsForLocation(tp *Tracepoint) error {
	locs, err := d.FindLocation(-1, 0, 0, tp.Location, true, nil)
	if err != nil {
		return fmt.Errorf("failed to get locations: %w", err)
	}

	for i, loc := range locs {
		bpi := &breakpointInstance{
			tracepoint: tp,
		}

		if err := bpi.prepare(i); err != nil {
			return err
		}

		bp := tp.Breakpoint
		bp.UserData = bpi
		bp.Addrs = loc.PCs
		bp.AddrPid = loc.PCPids

		if bpi.Breakpoint, err = d.CreateBreakpoint(&bp, tp.Location, nil, false); err != nil {
			return fmt.Errorf("failed to create breakpoint: %w", err)
		}

		fields := breakpointFields(bpi.Breakpoint)
		fields = append(fields, zap.Int("num", i))
		d.logger.Debug("Created new breakpoint for location", fields...)
	}

	return nil
}

func ptrace(request int, pid int, addr uintptr, data uintptr) error {
	if _, _, e1 := syscall.Syscall6(syscall.SYS_PTRACE, uintptr(request), uintptr(pid), addr, data, 0, 0); e1 != 0 {
		return e1
	}

	return nil
}
