package gont

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/service"
	"github.com/go-delve/delve/service/api"
	"github.com/go-delve/delve/service/dap"
	debug "github.com/go-delve/delve/service/debugger"
	"github.com/stv0g/gont/pkg/trace"
	"go.uber.org/zap"
)

type debuggerInstance struct {
	*debug.Debugger

	debugger   *Debugger
	cmd        *exec.Cmd
	stop       chan any
	listenAddr *net.TCPAddr
	config     service.Config

	logger *zap.Logger
}

func (d *Debugger) Start(c *exec.Cmd) error {
	var err error

	di := &debuggerInstance{
		cmd:      c,
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
		return fmt.Errorf("failed to create debugger: %w", err)
	}

	for _, tp := range d.Tracepoints {
		tp := tp
		if err := di.createBreakpoints(&tp); err != nil {
			return fmt.Errorf("failed to create breakpoints: %w", err)
		}
	}

	if len(d.Tracers) > 0 && len(d.Tracepoints) > 0 {
		for _, t := range d.Tracers {
			if err := t.start(); err != nil {
				return fmt.Errorf("failed to start tracer: %w", err)
			}
		}

		go di.run()
	} else if d.BreakOnEntry {
		d.logger.Info("Debugger started waiting for connection", zap.Any("addr", di.listenAddr))
	} else {
		if _, err := di.Command(&api.DebuggerCommand{Name: api.Continue}, nil); err != nil {
			return err
		}
	}

	d.instances[c.Process.Pid] = di

	return err
}

func (d *debuggerInstance) run() {
	for {
		s, err := d.Command(&api.DebuggerCommand{Name: api.Continue}, nil)
		if err != nil {
			d.logger.Error("Failed to continue", zap.Error(err))
			break
		}

		for _, bp := range s.WatchOutOfScope {
			d.logger.Debug("Watch out of scope", zap.Any("bp", bp))
		}

		reason := d.StopReason()

		d.logger.Debug("Stopped",
			zap.Any("reason", reason),
			zap.Bool("running", s.Running),
			zap.Bool("exited", s.Exited),
		)

		if s.Exited {
			d.logger.Info("Process exited")
			break
		}

		if s.Running {
			d.logger.Warn("Process is still running")
			continue
		}

		if reason == proc.StopBreakpoint || reason == proc.StopWatchpoint {
			d.handleBreakpoint(s.CurrentThread)
		}
	}
}

func (d *debuggerInstance) handleBreakpoint(thr *api.Thread) {
	// Creation of delayed watch points
	if bpi, ok := thr.Breakpoint.UserData.(*breakpointInstance); ok &&
		bpi.tracepoint.WatchExpr != "" && bpi.tracepoint.WatchType != 0 {
		if err := bpi.createWatchpoint(d, thr); err != nil {
			d.logger.Error("Failed to create watchpoint", zap.Error(err))
		}

		bpi.breakpoint.Disabled = true
		if err := d.AmendBreakpoint(bpi.breakpoint); err != nil {
			d.logger.Error("Failed to delete breakpoint", zap.Error(err))
		}
	}

	if tep, ok := thr.Breakpoint.UserData.(traceEventPoint); ok {
		te := tep.traceEvent(thr)
		for _, t := range d.debugger.Tracers {
			t.newEvent(te)
		}
	}
}

func (d *debuggerInstance) listen(addr *net.TCPAddr) {
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-d.stop:
				// We were supposed to exit, do nothing and return
				return
			default:
				panic(err)
			}
		}

		ds := dap.NewSession(conn, &dap.Config{
			Config:        &d.config,
			StopTriggered: make(chan struct{}),
		}, d.Debugger)
		go ds.ServeDAPCodec()
	}
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

	bp := tp.Breakpoint
	bp.UserData = bpi

	if bpi.breakpoint, err = d.CreateBreakpoint(&bp, "", nil, false); err != nil {
		return fmt.Errorf("failed to create breakpoint: %w", err)
	}

	d.logger.Debug("Created new breakpoint", breakpointFields(bpi.breakpoint)...)

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

		bp := tp.Breakpoint
		bp.UserData = bpi
		bp.Addrs = loc.PCs
		bp.AddrPid = loc.PCPids

		if tp.Message != "" {
			bpi.message, err = parseDebugMessage(tp.Message)
			if err != nil {
				return fmt.Errorf("failed to parse message: %w", err)
			}

			bp.Variables = append(bp.Variables, bpi.message.args...)
		}

		if bp.Name == "" {
			bp.Name = fmt.Sprintf("%s (%d)", tp.Location, i)
		}

		if bpi.breakpoint, err = d.CreateBreakpoint(&bp, tp.Location, nil, false); err != nil {
			return fmt.Errorf("failed to create breakpoint: %w", err)
		}

		d.logger.Debug("Created new breakpoint for location", breakpointFields(bpi.breakpoint)...)
	}

	return nil
}

type traceEventPoint interface {
	traceEvent(*api.Thread) trace.Event
}

type breakpointInstance struct {
	tracepoint *Tracepoint
	breakpoint *api.Breakpoint
	message    *debugMessage
}

type watchpointInstance struct {
	tracepoint *Tracepoint
	breakpoint *api.Breakpoint
}

func (bpi *breakpointInstance) createWatchpoint(d *debuggerInstance, thr *api.Thread) error {
	wp, err := d.CreateWatchpoint(thr.GoroutineID, 0, 0, bpi.tracepoint.WatchExpr, bpi.tracepoint.WatchType)
	if err != nil {
		return fmt.Errorf("failed to create watchpoint: %w", err)
	}

	wp.UserData = &watchpointInstance{
		breakpoint: wp,
		tracepoint: bpi.tracepoint,
	}

	if err := d.AmendBreakpoint(wp); err != nil {
		return fmt.Errorf("failed to amend watchpoint: %w", err)
	}

	d.logger.Debug("Created new watchpoint", breakpointFields(wp)...,
	)

	return nil
}

func (bpi *breakpointInstance) TraceEvent(t *api.Thread) trace.Event {
	var msg string
	if bpi.message == nil {
		msg = fmt.Sprintf("Hit breakpoint %d: %s", bpi.breakpoint.ID, bpi.breakpoint.Name)
	} else {
		msg = bpi.message.evaluate(t.BreakpointInfo)
	}

	return trace.Event{
		Timestamp: time.Now(),
		Type:      "breakpoint",
		Message:   msg,
		Line:      bpi.breakpoint.Line,
		File:      bpi.breakpoint.File,
		Function:  bpi.breakpoint.FunctionName,
		Data:      breakpointData(t.Breakpoint, t.BreakpointInfo),
	}
}

func (wpi *watchpointInstance) traceEvent(t *api.Thread) trace.Event {
	return trace.Event{
		Timestamp: time.Now(),
		Type:      "watchpoint",
		Message:   fmt.Sprintf("Hit watchpoint %d: %s", wpi.breakpoint.ID, wpi.breakpoint.Name),
		Data:      breakpointData(t.Breakpoint, t.BreakpointInfo),
	}
}

func ptrace(request int, pid int, addr uintptr, data uintptr) error {
	if _, _, e1 := syscall.Syscall6(syscall.SYS_PTRACE, uintptr(request), uintptr(pid), uintptr(addr), uintptr(data), 0, 0); e1 != 0 {
		return syscall.Errno(e1)
	}

	return nil
}

func breakpointFields(bp *api.Breakpoint) []zap.Field {
	fields := []zap.Field{}

	if bp.Name != "" {
		fields = append(fields, zap.String("name", bp.Name))
	}

	if bp.FunctionName != "" {
		fields = append(fields, zap.String("function", bp.FunctionName))
	}

	if bp.File != "" {
		fields = append(fields, zap.String("file", bp.File))
	}

	if bp.Line != 0 {
		fields = append(fields, zap.Int("line", bp.Line))
	}

	if bp.ID >= 0 {
		fields = append(fields, zap.Int("id", bp.ID))
	}

	if bp.Addr != 0 {
		fields = append(fields, zap.Uint64("addr", bp.Addr))
	}

	if len(bp.Addrs) > 0 {
		fields = append(fields, zap.Uint64s("addrs", bp.Addrs))
	}

	if bp.Cond != "" {
		fields = append(fields, zap.String("cond", bp.Cond))
	}

	if bp.WatchExpr != "" {
		fields = append(fields, zap.String("watch_expr", bp.WatchExpr))
	}

	return fields
}

func breakpointData(bp *api.Breakpoint, bpInfo *api.BreakpointInfo) any {
	data := map[string]any{
		"id":        bp.ID,
		"hit_count": bp.TotalHitCount,
	}

	if bp.Name != "" {
		data["name"] = bp.Name
	}

	if bpInfo != nil {
		data["info"] = breakpointInfoData(bpInfo)
	}

	switch bpi := bp.UserData.(type) {
	case *breakpointInstance:
		data["user"] = bpi.tracepoint.UserData
	case *watchpointInstance:
		data["user"] = bpi.tracepoint.UserData
	}

	return data
}

func breakpointInfoData(bpInfo *api.BreakpointInfo) any {
	data := map[string]any{}

	if vars := variableData(bpInfo.Variables); len(vars) > 0 {
		data["variables"] = vars
	}

	if args := variableData(bpInfo.Arguments); len(args) > 0 {
		data["arguments"] = args
	}

	if locals := variableData(bpInfo.Locals); len(locals) > 0 {
		data["locals"] = locals
	}

	if st := stacktraceData(bpInfo.Stacktrace); len(st) > 0 {
		data["stacktrace"] = st
	}

	return data
}

func variableData(varList []api.Variable) map[string]any {
	vars := map[string]any{}

	for _, v := range varList {
		vars[v.Name] = v.Value
	}

	return vars
}

func stacktraceData(fs []api.Stackframe) []any {
	ss := []any{}

	for _, f := range fs {
		s := map[string]any{
			"arguments": variableData(f.Arguments),
			"locals":    variableData(f.Locals),
			"file":      f.File,
			"line":      f.Line,
		}

		if name := f.Function.Name_; name != "" {
			s["function"] = name
		}

		ss = append(ss, s)
	}

	return ss
}
