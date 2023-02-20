package cmd

import (
	"bytes"

	g "github.com/stv0g/gont/pkg"
)

type StdoutBuffer struct {
	*bytes.Buffer
}

func (s StdoutBuffer) ApplyCmd(c *g.Cmd) {
	c.StdoutWriters = append(c.StdoutWriters, s)
}

func Stdout(b *bytes.Buffer) StdoutBuffer {
	return StdoutBuffer{b}
}

type StderrBuffer struct {
	*bytes.Buffer
}

func (s StderrBuffer) ApplyCmd(c *g.Cmd) {
	c.StderrWriters = append(c.StderrWriters, s)
}

func Stderr(b *bytes.Buffer) StderrBuffer {
	return StderrBuffer{b}
}

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
