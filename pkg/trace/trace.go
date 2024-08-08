// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

// The following functions are intended to by used used for instrumentation of Go code
// which is started by gont.Node.{Start,StartWith,Run}

var (
	ErrTracingAlreadyEnabled = errors.New("tracing already enabled")
	ErrTracingNotSupported   = errors.New("tracing not supported")
	ErrTracingNotRunning     = errors.New("tracing not running")
)

//nolint:gochecknoglobals
var (
	eventCallback EventCallback
	eventWriter   io.Writer
	eventFile     *os.File
)

func Start(bufsize int) error {
	if eventWriter != nil {
		return ErrTracingAlreadyEnabled
	}

	traceFileName := os.Getenv("GONT_TRACEFILE")
	if traceFileName == "" {
		return fmt.Errorf("%w: Missing GONT_TRACEFILE environment variable", ErrTracingNotSupported)
	}

	var err error
	if eventFile, err = os.OpenFile(traceFileName, os.O_WRONLY|os.O_APPEND, 0o300); err != nil {
		return err
	}

	if bufsize > 0 {
		eventWriter = bufio.NewWriterSize(eventWriter, bufsize)
	} else {
		eventWriter = eventFile
	}

	return nil
}

func StartWithCallback(cb EventCallback) {
	eventCallback = cb
}

func Stop() error {
	if eventWriter == nil {
		return ErrTracingNotRunning
	}

	if bufferedTraceWriter, ok := eventWriter.(*bufio.Writer); ok {
		bufferedTraceWriter.Flush()
	}

	if err := eventFile.Close(); err != nil {
		return fmt.Errorf("failed to close trace file: %w", err)
	}

	eventWriter = nil

	return nil
}

func With(cb func() error, bufsize int) error {
	if err := Start(bufsize); err != nil {
		return err
	}

	cbErr := cb()

	if err := Stop(); err != nil {
		return err
	}

	return cbErr
}

func trace(data any, msg string) error {
	if eventWriter == nil && eventCallback == nil {
		return nil
	}

	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(3, pc)
	f := runtime.FuncForPC(pc[0])

	e := Event{
		Type:      "tracepoint",
		PID:       os.Getpid(),
		Timestamp: time.Now(),
		Message:   strings.TrimSpace(msg),
		Data:      data,
	}

	e.Function = f.Name()
	e.File, e.Line = f.FileLine(pc[0])

	if eventCallback != nil {
		eventCallback(e)
	}

	if eventWriter != nil {
		if _, err := e.WriteTo(eventWriter); err != nil {
			return err
		}
	}

	return nil
}

func PrintWithData(data any, msg string) error {
	return trace(data, msg)
}

func PrintfWithData(data any, f string, a ...any) error {
	return trace(data, fmt.Sprintf(f, a...))
}

func Print(msg string) error {
	return trace(nil, msg)
}

func Printf(f string, a ...any) error {
	return trace(nil, fmt.Sprintf(f, a...))
}
