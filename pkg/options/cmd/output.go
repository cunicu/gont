// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"io"

	g "cunicu.li/gont/v2/pkg"
)

// StdoutWriter allows passing an io.Writer to which
// all outputs of the standard output of the sub-process
// are written.
type StdoutWriter struct {
	io.Writer
}

func (s StdoutWriter) ApplyCmd(c *g.Cmd) {
	c.StdoutWriters = append(c.StdoutWriters, s)
}

func Stdout(wr io.Writer) StdoutWriter {
	return StdoutWriter{wr}
}

// StderrWriter allows passing an io.Writer to which
// all outputs of the standard error output of the sub-process
// are written.
type StderrWriter struct {
	io.Writer
}

func (s StderrWriter) ApplyCmd(c *g.Cmd) {
	c.StderrWriters = append(c.StderrWriters, s)
}

func Stderr(wr io.Writer) StderrWriter {
	return StderrWriter{wr}
}

// CombinedWriter allows passing an io.Writer to which
// all outputs of the standard and error output of the sub-process
// are written.
type CombinedWriter struct {
	io.Writer
}

func (s CombinedWriter) ApplyCmd(c *g.Cmd) {
	c.StdoutWriters = append(c.StdoutWriters, s)
	c.StderrWriters = append(c.StderrWriters, s)
}

func Combined(wr io.Writer) CombinedWriter {
	return CombinedWriter{wr}
}
