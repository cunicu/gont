package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type Path string

func (p Path) Apply(c *exec.Cmd) {
	c.Path = string(p)
}

type Arg string

func (a Arg) Apply(c *exec.Cmd) {
	c.Args = append(c.Args, string(a))
}

type Args []string

func (as Args) Apply(c *exec.Cmd) {
	for _, a := range []string(as) {
		c.Args = append(c.Args, a)
	}
}

type Stdin struct {
	io.Reader
}

func (s Stdin) Apply(c *exec.Cmd) {
	c.Stdin = s.Reader
}

type Stdout struct {
	io.Writer
}

func (s Stdout) Apply(c *exec.Cmd) {
	c.Stdout = s.Writer
}

type Stderr struct {
	io.Writer
}

func (s Stderr) Apply(c *exec.Cmd) {
	c.Stderr = s.Writer
}

type ExtraFile os.File

func (e *ExtraFile) Apply(c *exec.Cmd) {
	c.ExtraFiles = append(c.ExtraFiles, (*os.File)(e))
}

type SysProcAttr syscall.SysProcAttr

func (s *SysProcAttr) Apply(c *exec.Cmd) {
	c.SysProcAttr = (*syscall.SysProcAttr)(s)
}

type Dir string

func (d Dir) Apply(c *exec.Cmd) {
	c.Dir = string(d)
}

type Env string

func (e Env) Apply(c *exec.Cmd) {
	c.Env = append(c.Env, string(e))
}

func EnvVar(k, v string) Env {
	return Env(fmt.Sprintf("%s=%s", k, v))
}
