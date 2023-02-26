// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/service/api"
	"github.com/stv0g/gont/pkg/trace"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals
var watchpointLoadConfig = proc.LoadConfig{
	MaxStringLen:       512,
	MaxArrayValues:     512,
	MaxStructFields:    32,
	FollowPointers:     true,
	MaxVariableRecurse: 3,
}

type traceEventPoint interface {
	traceEvent(*api.Thread, *debuggerInstance) trace.Event
}

type breakpointInstance struct {
	*api.Breakpoint

	// The tracepoint configuration used for creating the breakpoint
	tracepoint *Tracepoint

	// The watchpoint which has been created by this tracepoint
	watchpoint *watchpointInstance

	message *debugMessage
}

type watchpointInstance struct {
	*api.Breakpoint

	breakpoint *breakpointInstance

	hitCount atomic.Uint64
}

func (bpi *breakpointInstance) prepare(i int) error {
	var err error

	tp := bpi.tracepoint

	if tp.Name == "" {
		tp.Name = fmt.Sprintf("%s (%d)", tp.Location, i)
	}

	if tp.Message != "" {
		bpi.message, err = parseDebugMessage(tp.Message)
		if err != nil {
			return fmt.Errorf("failed to parse message: %w", err)
		}

		tp.Variables = append(tp.Variables, bpi.message.args...)
	}

	if tp.File != "" && !filepath.IsAbs(tp.File) {
		if tp.File, err = filepath.Abs(tp.File); err != nil {
			return err
		}
	}

	return nil
}

func (bpi *breakpointInstance) createWatchpoint(d *debuggerInstance, thr *api.Thread) error {
	var wpi *watchpointInstance
	var err error

	if wpi = bpi.watchpoint; wpi == nil {
		wpi = &watchpointInstance{
			breakpoint: bpi,
		}
		bpi.watchpoint = wpi
	}

	if wpi.Breakpoint, err = d.CreateWatchpoint(thr.GoroutineID, 0, 0, bpi.tracepoint.WatchExpr, bpi.tracepoint.WatchType); err != nil {
		return err
	}

	wp := wpi.Breakpoint
	wp.UserData = wpi
	if err := d.AmendBreakpoint(wp); err != nil {
		return fmt.Errorf("failed to amend watchpoint: %w", err)
	}

	d.logger.Debug("Created new watchpoint", breakpointFields(wp)...)

	return nil
}

func (bpi *breakpointInstance) traceEvent(t *api.Thread, d *debuggerInstance) trace.Event {
	var msg string
	if bpi.message == nil {
		msg = fmt.Sprintf("Hit breakpoint %d: %s", bpi.ID, bpi.Name)
	} else {
		msg = bpi.message.evaluate(t.BreakpointInfo.Variables)
	}

	e := trace.Event{
		Timestamp: time.Now(),
		Type:      "breakpoint",
		Message:   msg,
		Line:      t.Line,
		File:      t.File,
		Function:  t.Function.Name_,
	}

	e.Breakpoint, e.Data = breakpointData(t.Breakpoint, t.BreakpointInfo)

	return e
}

func (wpi *watchpointInstance) traceEvent(t *api.Thread, d *debuggerInstance) trace.Event {
	e := trace.Event{
		Timestamp: time.Now(),
		Type:      "watchpoint",
		File:      t.File,
		Line:      t.Line,
		Function:  t.Function.Name_,
		Data:      wpi.breakpoint,
	}

	e.Breakpoint, e.Data = breakpointData(t.Breakpoint, t.BreakpointInfo)
	e.Breakpoint.Variables = wpi.evaluate(t, d)

	e.Breakpoint.TotalHitCount = wpi.hitCount.Add(1)

	if msg := wpi.breakpoint.message; msg != nil {
		e.Message = msg.evaluate(e.Breakpoint.Variables)
	} else {
		e.Message = fmt.Sprintf("Hit watchpoint %d: %s", wpi.ID, wpi.Name)
	}

	return e
}

func (wpi *watchpointInstance) evaluate(t *api.Thread, d *debuggerInstance) []api.Variable {
	tp := wpi.breakpoint

	vars := []api.Variable{}

	for _, v := range tp.Variables {
		ev, err := d.EvalVariableInScope(t.GoroutineID, 0, 0, v, watchpointLoadConfig)
		if err != nil {
			continue
		}

		vars = append(vars, *api.ConvertVar(ev))
	}

	return vars
}

func (wpi *watchpointInstance) wentOutOfScope(d *debuggerInstance) error {
	d.logger.Debug("Watch went out of scope", zap.Any("wp", wpi))

	bp := wpi.breakpoint.Breakpoint

	bp.Disabled = false
	if err := d.AmendBreakpoint(bp); err != nil {
		return fmt.Errorf("failed to reenable watchpoint: %w", err)
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

func breakpointData(bp *api.Breakpoint, bpInfo *api.BreakpointInfo) (*trace.Breakpoint, any) {
	b := &trace.Breakpoint{
		ID:            bp.ID,
		Name:          bp.Name,
		TotalHitCount: bp.TotalHitCount,
		HitCount:      bp.HitCount,
		Variables:     bpInfo.Variables,
		Arguments:     bpInfo.Arguments,
		Locals:        bpInfo.Locals,
		Stacktrace:    bpInfo.Stacktrace,
	}

	switch i := bp.UserData.(type) {
	case *breakpointInstance:
		return b, i.tracepoint.UserData
	case *watchpointInstance:
		return b, i.breakpoint.tracepoint.UserData
	}

	return b, nil
}
