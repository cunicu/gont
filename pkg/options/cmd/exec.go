// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

// Path is the path to the executable of the sub-process.
type Path string

func (p Path) ApplyExecCmd(c *exec.Cmd) {
	c.Path = string(p)
}

// Arg appends a single argument to the list of arguments
//
// This option can be specified multiple times to provide
// multiple arguments. Please keep in mind that the order
// is maintained.
type Arg string

func (a Arg) ApplyExecCmd(c *exec.Cmd) {
	c.Args = append(c.Args, string(a))
}

// Arguments appends multiple arguments to the list of arguments
type Arguments []string

func (as Arguments) ApplyExecCmd(c *exec.Cmd) {
	c.Args = append(c.Args, as...)
}

// Stdin set the standard input reader
type Stdin struct {
	io.Reader
}

func (s Stdin) ApplyExecCmd(c *exec.Cmd) {
	c.Stdin = s.Reader
}

// ExtraFile appends an additional file to the list of file descriptors
// which are passed on to the sub-process.
type ExtraFile os.File

func (e *ExtraFile) ApplyExecCmd(c *exec.Cmd) {
	c.ExtraFiles = append(c.ExtraFiles, (*os.File)(e))
}

// SysProcAttr allows to set system specific process attributes.
type SysProcAttr syscall.SysProcAttr

func (s *SysProcAttr) ApplyExecCmd(c *exec.Cmd) {
	c.SysProcAttr = (*syscall.SysProcAttr)(s)
}

// Dir sets the working directory.
type Dir string

func (d Dir) ApplyExecCmd(c *exec.Cmd) {
	c.Dir = string(d)
}

// Env appends additional environment variables.
type Env string

func (e Env) ApplyExecCmd(c *exec.Cmd) {
	c.Env = append(c.Env, string(e))
}

// EnvVar appends a key-value paired environment variable
func EnvVar(k, v string) Env {
	return Env(fmt.Sprintf("%s=%s", k, v))
}

// Envs appends additional environment variables.
type Envs []string

func (e Envs) ApplyExecCmd(c *exec.Cmd) {
	c.Env = append(c.Env, e...)
}
