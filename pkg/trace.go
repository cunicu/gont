package gont

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/stv0g/gont/internal/prque"
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
	stop    chan any
	queue   *prque.PriorityQueue
	logger  *zap.Logger
}

func NewTracer() *Tracer {
	return &Tracer{
		stop:   make(chan any),
		queue:  prque.New(),
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

	go t.writeEvents()

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

func (t *Tracer) Flush() error {
	for t.queue.Len() > 0 {
		p := t.queue.Pop().(trace.Event)

		t.writeEvent(p)
	}

	return nil
}

func (t *Tracer) Close() error {
	close(t.stop)

	if err := t.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

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
	if len(t.Channels)+len(t.Callbacks)+len(t.files) > 0 {
		t.queue.Push(e)
	}

	for _, ps := range t.packetSources {
		ps.SourceTracepoint(e)
	}
}

func (t *Tracer) writeEvent(e trace.Event) error {
	for _, ch := range t.Channels {
		ch <- e
	}

	for _, cb := range t.Callbacks {
		cb(e)
	}

	for _, file := range t.files {
		if err := e.Log(file); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tracer) writeEvents() {
	tickerEvents := time.NewTicker(1 * time.Second)

out:
	for {
		select {
		case now := <-tickerEvents.C:
			for {
				if t.queue.Len() < 1 {
					break
				}

				oldest := t.queue.Oldest()
				oldestAge := now.Sub(oldest)
				if oldestAge < 1*time.Second {
					break
				}

				e := t.queue.Pop().(trace.Event)

				if err := t.writeEvent(e); err != nil {
					t.logger.Error("Failed to handle event. Stop tracing...", zap.Error(err))
					break out
				}
			}

		case <-t.stop:
			return
		}
	}
}
