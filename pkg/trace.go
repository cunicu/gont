package gont

import (
	"errors"
	"io"
	"os"

	"github.com/stv0g/gont/pkg/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"
)

type TracepointCallbackFunc func(tp trace.Event)

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
	Channels  []chan trace.Event
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
			tp := trace.Event{}
			if _, err := tp.ReadFrom(rd); err != nil {
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
