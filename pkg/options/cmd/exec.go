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

type Path string

func (p Path) ApplyExecCmd(c *exec.Cmd) {
	c.Path = string(p)
}

type Arg string

func (a Arg) ApplyExecCmd(c *exec.Cmd) {
	c.Args = append(c.Args, string(a))
}

type Arguments []string

func (as Arguments) ApplyExecCmd(c *exec.Cmd) {
	for _, a := range []string(as) {
		c.Args = append(c.Args, a)
	}
}

type Stdin struct {
	io.Reader
}

func (s Stdin) ApplyExecCmd(c *exec.Cmd) {
	c.Stdin = s.Reader
}

type ExtraFile os.File

func (e *ExtraFile) ApplyExecCmd(c *exec.Cmd) {
	c.ExtraFiles = append(c.ExtraFiles, (*os.File)(e))
}

type SysProcAttr syscall.SysProcAttr

func (s *SysProcAttr) ApplyExecCmd(c *exec.Cmd) {
	c.SysProcAttr = (*syscall.SysProcAttr)(s)
}

type Dir string

func (d Dir) ApplyExecCmd(c *exec.Cmd) {
	c.Dir = string(d)
}

type Env string

func (e Env) ApplyExecCmd(c *exec.Cmd) {
	c.Env = append(c.Env, string(e))
}

func EnvVar(k, v string) Env {
	return Env(fmt.Sprintf("%s=%s", k, v))
}
