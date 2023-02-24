package cmd

import (
	"bytes"

	g "github.com/stv0g/gont/pkg"
)

// StdoutBuffer allows passing a bytes.Buffer which will record
// all outputs of the standard output of the sub-process.
type StdoutBuffer struct {
	*bytes.Buffer
}

func (s StdoutBuffer) ApplyCmd(c *g.Cmd) {
	c.StdoutWriters = append(c.StdoutWriters, s)
}

func Stdout(b *bytes.Buffer) StdoutBuffer {
	return StdoutBuffer{b}
}

// StderrBuffer allows passing a bytes.Buffer which will record
// all outputs of the standard error output of the sub-process.
type StderrBuffer struct {
	*bytes.Buffer
}

func (s StderrBuffer) ApplyCmd(c *g.Cmd) {
	c.StderrWriters = append(c.StderrWriters, s)
}

func Stderr(b *bytes.Buffer) StderrBuffer {
	return StderrBuffer{b}
}

// CombinedBuffer allows passing a bytes.Buffer which will record
// the combined outputs of the standard and standard error output
// of the sub-process.
type CombinedBuffer struct {
	*bytes.Buffer
}

func (s CombinedBuffer) ApplyCmd(c *g.Cmd) {
	c.StdoutWriters = append(c.StdoutWriters, s)
	c.StderrWriters = append(c.StderrWriters, s)
}

func Combined(b *bytes.Buffer) CombinedBuffer {
	return CombinedBuffer{b}
}
