// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"encoding/json"
	"io"
	"reflect"
	"time"

	"github.com/fxamacker/cbor/v2"
)

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
	Timestamp time.Time `cbor:"time" json:"time"`
	Type      string    `cbor:"type" json:"type"`
	Level     uint8     `cbor:"lvl,omitempty" json:"lvl,omitempty"`
	Message   string    `cbor:"msg,omitempty" json:"msg,omitempty"`
	Source    string    `cbor:"src,omitempty" json:"src,omitempty"`
	PID       int       `cbor:"pid,omitempty" json:"pid,omitempty"`
	Function  string    `cbor:"func,omitempty" json:"func,omitempty"`
	File      string    `cbor:"file,omitempty" json:"file,omitempty"`
	Line      int       `cbor:"line,omitempty" json:"line,omitempty"`
	Args      []any     `cbor:"args,omitempty" json:"args,omitempty"`
	Data      any       `cbor:"data,omitempty" json:"data,omitempty"`
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

func (e *Event) Log(wr io.Writer) error {
	return json.NewEncoder(wr).Encode(e)
}

func (e *Event) Marshal() ([]byte, error) {
	return em.Marshal(e)
}

func (e *Event) Unmarshal(b []byte) error {
	return dm.Unmarshal(b, e)
}
