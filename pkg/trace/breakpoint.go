// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-delve/delve/service/api"
)

type Stacktrace struct {
	Arguments []api.Variable `cbor:"arguments,omitempty" json:"arguments,omitempty"`
	Locals    []api.Variable `cbor:"locals,omitempty" json:"locals,omitempty"`
	Function  string         `cbor:"function,omitempty" json:"function,omitempty"`
	File      string         `cbor:"file,omitempty" json:"file,omitempty"`
	Line      int            `cbor:"line,omitempty" json:"line,omitempty"`
}

//nolint:tagliatelle
type Breakpoint struct {
	ID            int               `cbor:"id" json:"id"`
	Name          string            `cbor:"name,omitempty" json:"name,omitempty"`
	TotalHitCount uint64            `cbor:"total_hit_count,omitempty" json:"total_hit_count,omitempty"`
	HitCount      map[string]uint64 `cbor:"hit_count,omitempty" json:"hit_count,omitempty"`
	Variables     []api.Variable    `cbor:"vars,omitempty" json:"vars,omitempty"`
	Arguments     []api.Variable    `cbor:"args,omitempty" json:"args,omitempty"`
	Locals        []api.Variable    `cbor:"locals,omitempty" json:"locals,omitempty"`
	Stacktrace    []api.Stackframe  `cbor:"stack,omitempty" json:"stack,omitempty"`
}

func (b *Breakpoint) Variable(name string) string {
	for _, v := range b.Variables {
		if v.Name == name {
			return v.Value
		}
	}

	return ""
}

func (b *Breakpoint) Argument(name string) string {
	for _, v := range b.Arguments {
		if v.Name == name {
			return v.Value
		}
	}

	return ""
}

func (b *Breakpoint) Local(name string) string {
	for _, v := range b.Locals {
		if v.Name == name {
			return v.Value
		}
	}

	return ""
}

func fprintStacktrace(w io.Writer, st []api.Stackframe) {
	if len(st) == 0 {
		return
	}

	fmt.Fprintln(w, "  Stacktrace:")
	for _, s := range st {
		argStrs := []string{}
		for _, arg := range s.Arguments {
			argStrs = append(argStrs, arg.SinglelineString())
		}

		fmt.Fprintf(w, "    %s(%s)\n", s.Function.Name(), strings.Join(argStrs, ", "))
		fmt.Fprintf(w, "      %s:%d\n", s.File, s.Line)
	}
}

func fprintVariables(w io.Writer, indent, title string, vars []api.Variable) {
	if len(vars) == 0 {
		return
	}

	fmt.Fprintf(w, indent+"%s:\n", title)
	for _, v := range vars {
		vs := v.MultilineString("    ", "")
		fmt.Fprintf(w, indent+"  %s: %s\n", v.Name, vs)
	}
}
