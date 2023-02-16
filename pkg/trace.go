package gont

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/fxamacker/cbor/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"
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

type TracepointCallbackFunc func(tp Tracepoint)

// Options

type TraceOption interface {
	Apply(n *Tracer)
}

func (c *Tracer) Apply(n *BaseNode) {
	n.Tracer = c
}

type Tracer struct {
	// Output options
	Files     []*os.File
	Filenames []string
	Channels  []chan Tracepoint
	Callbacks []TracepointCallbackFunc
	Captures  []*Capture
}

func (t *Tracer) Close() error {
	return nil
}

func (t *Tracer) Pipe() (*os.File, error) {
	rd, wr, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	// Filenames
	files := []*os.File{}
	for _, fn := range t.Filenames {
		f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	zap.L().Info("Opened pipe files")

	// Captures
	tpss := []*tracepointPacketSource{}
	for _, c := range t.Captures {
		_, tps, err := c.startTrace()
		if err != nil {
			return nil, err
		}

		tpss = append(tpss, tps)
	}

	go func() {
		for {
			tp := Tracepoint{}
			if err := tp.ReadFrom(rd); err != nil {
				if errors.Is(err, io.EOF) {
					break
				} else {
					zap.L().Warn("Failed to read tracepoint from log", zap.Error(err))
					continue
				}
			}

			// TODO: Remove
			wr := &zapio.Writer{
				Log:   zap.L(),
				Level: zapcore.InfoLevel,
			}
			tp.Log(wr)

			for _, file := range files {
				tp.Log(file)
			}

			for _, ch := range t.Channels {
				ch <- tp
			}

			for _, cb := range t.Callbacks {
				cb(tp)
			}

			for _, tps := range tpss {
				tps.SourceTracepoint(tp)
			}
		}

		for _, tps := range tpss {
			tps.Close()
		}

		for _, file := range files {
			file.Close()
		}

		zap.L().Info("Closed pipe files")
	}()

	return wr, nil
}

type Tracepoint struct {
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

func (t *Tracepoint) WriteTo(wr io.Writer) error {
	return em.NewEncoder(wr).Encode(t)
}

func (t *Tracepoint) ReadFrom(rd io.Reader) error {
	return dm.NewDecoder(rd).Decode(t)
}

func (t *Tracepoint) Log(wr io.Writer) error {
	return json.NewEncoder(wr).Encode(t)
}
