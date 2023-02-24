package debug

import (
	"github.com/go-delve/delve/service/api"
	g "github.com/stv0g/gont/pkg"
)

// User defined name of the breakpoint.
type Name string

func (n Name) ApplyBreakpoint(b *g.Tracepoint) {
	b.Name = string(n)
}

// Location is a Delve locspec
// See: https://github.com/go-delve/delve/blob/master/Documentation/cli/locspec.md
type Location string

func (l Location) ApplyBreakpoint(b *g.Tracepoint) {
	b.Location = string(l)
}

// FunctionNameRegex is a regular expression which matches
// against all available functions in the debugged executable.
// A dedicated tracepoint will be created for each matched function.
func FunctionNameRegex(expr string) Location {
	return Location("/" + expr + "/")
}

// Address is deprecated, use Addrs.
type Address uint64

func (a Address) ApplyBreakpoint(b *g.Tracepoint) {
	b.Addr = uint64(a)
}

// File is the source file for the breakpoint.
type File string

func (f File) ApplyBreakpoint(b *g.Tracepoint) {
	b.File = string(f)
}

// Line is a line in File for the breakpoint.
type Line int

func (l Line) ApplyBreakpoint(b *g.Tracepoint) {
	b.Line = int(l)
}

// FunctionName is the name of the function at the current breakpoint, and
// may not always be available.
type FunctionName string

func (fn FunctionName) ApplyBreakpoint(b *g.Tracepoint) {
	b.FunctionName = string(fn)
}

// Breakpoint condition.
type Condition string

func (c Condition) ApplyBreakpoint(b *g.Tracepoint) {
	b.Cond = string(c)
}

// Breakpoint hit count condition.
// Supported hit count conditions are "NUMBER" and "OP NUMBER".
type HitCondition string

func (hc HitCondition) ApplyBreakpoint(b *g.Tracepoint) {
	b.HitCond = string(hc)
}

// HitConditionPerGoroutine use per goroutine hitcount as HitCond operand, instead of total hitcount.
type HitConditionPerGoroutine bool

func (hc HitConditionPerGoroutine) ApplyBreakpoint(b *g.Tracepoint) {
	b.HitCondPerG = bool(hc)
}

// Tracepoint flag, signifying this is a tracepoint.
type Tracepoint bool

func (tp Tracepoint) ApplyBreakpoint(b *g.Tracepoint) {
	b.Tracepoint = bool(tp)
}

// TraceReturn flag signifying this is a breakpoint set at a return
// statement in a traced function.
type TraceReturn bool

func (tr TraceReturn) ApplyBreakpoint(b *g.Tracepoint) {
	b.TraceReturn = bool(tr)
}

// Retrieve goroutine information
type Goroutine bool

func (g Goroutine) ApplyBreakpoint(b *g.Tracepoint) {
	b.Goroutine = bool(g)
}

// Stack frames to retrieve
type Stacktrace int

func (s Stacktrace) ApplyBreakpoint(b *g.Tracepoint) {
	b.Stacktrace = int(s)
}

// Expressions to evaluate
type Variable string

func (v Variable) ApplyBreakpoint(b *g.Tracepoint) {
	b.Variables = append(b.Variables, string(v))
}

// WatchExpression is the expression used to create this watchpoint
type WatchExpression struct {
	Expr string
	Type api.WatchType
}

func (w WatchExpression) ApplyBreakpoint(b *g.Tracepoint) {
	b.WatchExpr = w.Expr
	b.WatchType = w.Type
}

func Watch(expr string, typ api.WatchType) WatchExpression {
	return WatchExpression{
		Expr: expr,
		Type: typ,
	}
}

type VerboseDescription []string

func (v VerboseDescription) ApplyBreakpoint(b *g.Tracepoint) {
	b.VerboseDescr = []string(v)
}

func Description(strs ...string) VerboseDescription {
	return VerboseDescription(strs)
}

// Disabled flag, signifying the state of the breakpoint
type Disabled bool

func (d Disabled) ApplyBreakpoint(b *g.Tracepoint) {
	b.Disabled = bool(d)
}

// UserData carries arbitrary data which will be attached
// to the trace events.
type UserData struct {
	Data any
}

func (u UserData) ApplyBreakpoint(b *g.Tracepoint) {
	b.UserData = any(u.Data)
}

func Data(d any) UserData {
	return UserData{
		Data: d,
	}
}

func LoadConfig(opts ...LoadConfigOption) api.LoadConfig {
	lc := api.LoadConfig{}

	for _, opt := range opts {
		opt.ApplyLoadConfig(&lc)
	}

	return lc
}

// LoadArguments requests loading function arguments when the breakpoint is hit
type LoadArgumentsConfig api.LoadConfig

func (l LoadArgumentsConfig) ApplyBreakpoint(b *g.Tracepoint) {
	b.LoadArgs = (*api.LoadConfig)(&l)
}

func LoadArguments(opts ...LoadConfigOption) LoadArgumentsConfig {
	return LoadArgumentsConfig(LoadConfig(opts...))
}

// LoadLocals requests loading function locals when the breakpoint is hit
type LoadLocalsConfig api.LoadConfig

func (l LoadLocalsConfig) ApplyBreakpoint(b *g.Tracepoint) {
	b.LoadLocals = (*api.LoadConfig)(&l)
}

func LoadLocals(opts ...LoadConfigOption) LoadLocalsConfig {
	return LoadLocalsConfig(LoadConfig(opts...))
}

// Message overwrite the message of the trace events which
// are produced when the tracepoint is hit.
//
// The message can contain placeholders which are substituted
// with variables in the scope of the breakpoint.
//
// Example: "The variable myVar has currently the value {myVar}"
type Message string

func (m Message) ApplyBreakpoint(b *g.Tracepoint) {
	b.Message = string(m)
}
