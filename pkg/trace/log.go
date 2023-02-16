package trace

import (
	"os"
	"strings"
	"time"

	g "github.com/stv0g/gont/pkg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Log() zap.Option {
	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, &traceCore{})
	})
}

type traceCore struct {
	fields []zapcore.Field
}

func (c *traceCore) Enabled(lvl zapcore.Level) bool {
	return traceFile != nil
}

func (c *traceCore) With(fields []zapcore.Field) zapcore.Core {
	return &traceCore{
		fields: append(c.fields, fields...),
	}
}

func (c *traceCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(e.Level) {
		ce.AddCore(e, c)
	}

	return ce
}

func (c *traceCore) Write(e zapcore.Entry, fields []zapcore.Field) error {
	if traceWriter == nil {
		return nil
	}

	enc := zapcore.NewMapObjectEncoder()

	for _, f := range c.fields {
		f.AddTo(enc)
	}
	for _, f := range fields {
		f.AddTo(enc)
	}

	t := g.Tracepoint{
		Type:      "log",
		PID:       os.Getpid(),
		Timestamp: time.Now(),
		Message:   strings.TrimSpace(e.Message),
		Source:    e.LoggerName,
		Level:     uint8(e.Level) + 2, // 0 -> omitempty
		Function:  e.Caller.Function,
		Line:      e.Caller.Line,
		File:      e.Caller.File,
		Data:      enc.Fields,
	}

	return t.WriteTo(traceWriter)
}

func (c *traceCore) Sync() error {
	if traceFile != nil {
		return traceFile.Sync()
	}
	return nil
}
