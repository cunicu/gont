// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/fxamacker/cbor/v2"
	"go.uber.org/zap/zapcore"
)

//nolint:gochecknoglobals
var (
	dm cbor.DecMode
	em cbor.EncMode
)

func init() {
	dm, _ = cbor.DecOptions{
		DefaultMapType: reflect.TypeOf(map[string]any{}),
	}.DecMode()

	em, _ = cbor.EncOptions{
		Time: cbor.TimeUnixMicro,
	}.EncMode()
}

type EventCallback func(e Event)

type Event struct {
	Timestamp  time.Time   `cbor:"time" json:"time"`
	Type       string      `cbor:"type" json:"type"`
	Level      uint8       `cbor:"lvl,omitempty" json:"lvl,omitempty"`
	Message    string      `cbor:"msg,omitempty" json:"msg,omitempty"`
	Source     string      `cbor:"src,omitempty" json:"src,omitempty"`
	PID        int         `cbor:"pid,omitempty" json:"pid,omitempty"`
	Function   string      `cbor:"func,omitempty" json:"func,omitempty"`
	File       string      `cbor:"file,omitempty" json:"file,omitempty"`
	Line       int         `cbor:"line,omitempty" json:"line,omitempty"`
	Breakpoint *Breakpoint `cbor:"breakpoint,omitempty" json:"breakpoint,omitempty"`
	Data       any         `cbor:"data,omitempty" json:"data,omitempty"`
}

func (e Event) Time() time.Time {
	return e.Timestamp
}

func (e *Event) WriteTo(wr io.Writer) (int64, error) {
	return 0, em.NewEncoder(wr).Encode(e)
}

func (e *Event) ReadFrom(rd io.Reader) (int64, error) {
	return 0, dm.NewDecoder(rd).Decode(e)
}

func (e *Event) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Event) Marshal() ([]byte, error) {
	return em.Marshal(e)
}

func (e *Event) Unmarshal(b []byte) error {
	return dm.Unmarshal(b, e)
}

func (e *Event) Fprint(w io.Writer) {
	indent := "  "

	ts := e.Timestamp.Format("15:04:05.999999")
	marker := strings.Repeat("-", 76-len(ts)-len(e.Message))

	fmt.Fprintf(w, "%s: %s %s\n", ts, e.Message, marker)
	fmt.Fprintf(w, indent+"Type:       %s\n", e.Type)

	if e.Level > 0 {
		lvl := zapcore.Level(e.Level - 2)
		fmt.Fprintf(w, indent+"Level:      %s\n", lvl.String())
	}

	if e.PID > 0 {
		fmt.Fprintf(w, indent+"PID:        %d\n", e.PID)
	}

	if e.Source != "" {
		fmt.Fprintf(w, indent+"Source:     %s\n", e.Source)
	}

	if e.Function != "" {
		fmt.Fprintf(w, indent+"Function:   %s\n", e.Function)
	}

	if e.File != "" {
		fmt.Fprintf(w, indent+"File/Line:  %s:%d\n", e.File, e.Line)
	}

	if e.Data != nil {
		cs := spew.ConfigState{Indent: indent}
		fmt.Fprintf(w, indent+"Data:       %v\n", cs.NewFormatter(e.Data))
	}

	if bp := e.Breakpoint; bp != nil {
		if bp.Name != "" {
			fmt.Fprintf(w, indent+"Breakpoint: %s (%d)\n", bp.Name, bp.ID)
		} else {
			fmt.Fprintf(w, indent+"Breakpoint: %d\n", bp.ID)
		}
		fmt.Fprintf(w, indent+"Hit count:  %d\n", bp.TotalHitCount)

		fprintVariables(w, indent, "Arguments", bp.Arguments)
		fprintVariables(w, indent, "Locals", bp.Locals)
		fprintVariables(w, indent, "Variables", bp.Variables)
		fprintStacktrace(w, bp.Stacktrace)
	}
}
