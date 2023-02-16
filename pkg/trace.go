package gont

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/stv0g/gont/pkg/trace"
	"go.uber.org/zap"
)

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
	Callbacks []trace.EventCallback
	Captures  []*Capture

	// Outputs
	files         []*os.File
	packetSources []*traceEventPacketSource

	started bool

	logger *zap.Logger
}

func NewTracer() *Tracer {
	return &Tracer{
		logger: zap.L().Named("tracer"),
	}
}

func (t *Tracer) Start() error {
	// Filenames
	for _, fn := range t.Filenames {
		f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		t.files = append(t.files, f)
	}

	// Captures
	for _, c := range t.Captures {
		_, ps, err := c.startTrace()
		if err != nil {
			return fmt.Errorf("failed to start capturing traces: %w", err)
		}

		t.packetSources = append(t.packetSources, ps)
	}

	t.started = true

	return nil
}

func (t *Tracer) StartLocal() error {
	if !t.started {
		if err := t.Start(); err != nil {
			return err
		}
	}

	trace.StartWithCallback(t.newEvent)

	return nil
}

func (t *Tracer) Close() error {
	for _, tps := range t.packetSources {
		if err := tps.Close(); err != nil {
			return err
		}
	}

	for _, file := range t.files {
		if err := file.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tracer) Pipe() (*os.File, error) {
	if !t.started {
		if err := t.Start(); err != nil {
			return nil, err
		}
	}

	rd, wr, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			e := trace.Event{}
			if _, err := e.ReadFrom(rd); err != nil {
				if errors.Is(err, io.EOF) {
					break
				} else {
					t.logger.Warn("Failed to read tracepoint from log", zap.Error(err))
					continue
				}
			}

			t.newEvent(e)
		}
	}()

	return wr, nil
}

func (t *Tracer) newEvent(e trace.Event) {
	for _, file := range t.files {
		e.Log(file)
	}

	for _, ch := range t.Channels {
		ch <- e
	}

	for _, cb := range t.Callbacks {
		cb(e)
	}

	for _, ps := range t.packetSources {
		ps.SourceTracepoint(e)
	}
}
