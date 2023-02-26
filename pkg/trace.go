// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

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
	ApplyTracer(t *Tracer)
}

func (t *Tracer) ApplyNetwork(n *Network) {
	n.Tracer = t
}

func (t *Tracer) ApplyBaseNode(n *BaseNode) {
	n.Tracer = t
}

func (t *Tracer) ApplyCmd(c *Cmd) {
	c.Tracer = t
}

func (t *Tracer) ApplyDebugger(d *Debugger) {
	d.Tracers = append(d.Tracers, t)
}

type Tracer struct {
	// Output options
	Files     []*os.File
	Filenames []string
	Channels  []chan trace.Event
	Callbacks []trace.EventCallback
	Captures  []*Capture

	closables     []io.Closer
	files         []*os.File
	packetSources []*traceEventPacketSource

	stop   chan any
	queue  *prque.PriorityQueue
	logger *zap.Logger
}

func NewTracer(opts ...TraceOption) *Tracer {
	t := &Tracer{
		queue:  prque.New(),
		logger: zap.L().Named("tracer"),
	}

	for _, opt := range opts {
		opt.ApplyTracer(t)
	}

	return t
}

func (t *Tracer) start() error {
	// Files
	t.files = t.Files

	// Filenames
	for _, filename := range t.Filenames {
		// TODO: It would be nice of we can use the same kind of templating
		//       for the filename as in Capture filenames. However, its tricky to
		//       include the PID into the template :(
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}

		t.files = append(t.files, file)
		t.closables = append(t.closables, file)
	}

	// Captures
	for _, c := range t.Captures {
		_, ps, err := c.startTrace()
		if err != nil {
			return fmt.Errorf("failed to start capturing traces: %w", err)
		}

		t.packetSources = append(t.packetSources, ps)
		t.closables = append(t.closables, ps)
	}

	t.stop = make(chan any)

	go t.writeEvents()

	return nil
}

func (t *Tracer) Start() error {
	if t.stop == nil {
		if err := t.start(); err != nil {
			return err
		}
	}

	trace.StartWithCallback(t.newEvent)

	return nil
}

func (t *Tracer) Flush() error {
	for t.queue.Len() > 0 {
		p := t.queue.Pop().(trace.Event)

		if err := t.writeEvent(p); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tracer) Close() error {
	if t.stop == nil {
		return nil // not started
	}

	close(t.stop)

	if err := t.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	for _, closable := range t.closables {
		if err := closable.Close(); err != nil {
			return fmt.Errorf("failed to close: %w", err)
		}
	}

	return nil
}

func (t *Tracer) Pipe() (*os.File, error) {
	if t.stop == nil {
		if err := t.start(); err != nil {
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
		if b, err := e.MarshalJSON(); err != nil {
			return err
		} else if _, err := file.Write(b); err != nil {
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
