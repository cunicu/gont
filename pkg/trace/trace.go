package trace

import (
	"bufio"
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
	traceWriter io.Writer
	traceFile   *os.File
)

func Start(bufsize int) error {
	if traceWriter != nil {
		return fmt.Errorf("tracing already enabled")
	}

	traceFileName := os.Getenv("GONT_TRACEFILE")
	if traceFileName == "" {
		return fmt.Errorf("tracing not supported. Missing GONT_TRACEFILE environment variable")
	}

	var err error
	if traceFile, err = os.OpenFile(traceFileName, os.O_WRONLY|os.O_APPEND, 0o300); err != nil {
		return err
	}

	if bufsize > 0 {
		traceWriter = bufio.NewWriterSize(traceWriter, bufsize)
	} else {
		traceWriter = traceFile
	}

	return nil
}

func Stop() error {
	if traceWriter == nil {
		return fmt.Errorf("tracing not running")
	}

	if bufferedTraceWriter, ok := traceWriter.(*bufio.Writer); ok {
		bufferedTraceWriter.Flush()
	}

	if err := traceFile.Close(); err != nil {
		return fmt.Errorf("failed to close trace file: %w", err)
	}

	traceWriter = nil

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
	if traceWriter == nil {
		return nil
	}

	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(3, pc)
	f := runtime.FuncForPC(pc[0])

	t := Event{
		Type:      "tracepoint",
		PID:       os.Getpid(),
		Timestamp: time.Now(),
		Message:   strings.TrimSpace(msg),
		Data:      data,
	}

	t.Function = f.Name()
	t.File, t.Line = f.FileLine(pc[0])

	_, err := t.WriteTo(traceWriter)
	return err
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
