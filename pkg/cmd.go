// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/gopacket/gopacket/pcapgo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

type ExecCmdOption interface {
	ApplyExecCmd(*exec.Cmd)
}

type CmdOption interface {
	ApplyCmd(*Cmd)
}

type Cmd struct {
	*exec.Cmd

	// Options
	Tracer        *Tracer
	RedirectToLog bool

	StdoutWriters []io.Writer
	StderrWriters []io.Writer

	stdoutPipe io.ReadCloser
	stderrPipe io.ReadCloser

	node   *BaseNode
	logger *zap.Logger
}

func (n *BaseNode) Command(name string, args ...any) *Cmd {
	c := &Cmd{
		node:          n,
		StdoutWriters: []io.Writer{},
		StderrWriters: []io.Writer{},
	}
	strArgs := []string{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case ExecCmdOption:
			arg.ApplyExecCmd(c.Cmd)
		case CmdOption:
			arg.ApplyCmd(c)
		default:
			if strArg, ok := stringifyArg(arg); ok {
				strArgs = append(strArgs, strArg)
			}
		}
	}

	c.logger = n.logger.Named("cmd").With(
		zap.String("path", name),
		zap.Strings("args", strArgs),
	)

	// Actual namespace switching is done similar to Docker's reexec
	// in a forked version of ourself by passing all required details
	// in environment variables.

	c.Cmd = exec.Command(name, strArgs...)

	if !c.node.NsHandle.Equal(c.node.network.HostNode.NsHandle) {
		if c.node.ExistingDockerContainer == "" {
			c.Path = "/proc/self/exe"
			c.Env = append(os.Environ(),
				"GONT_UNSHARE=exec",
				"GONT_NODE="+c.node.name,
				"GONT_NETWORK="+c.node.network.Name)
		} else {
			c.Path = "/usr/bin/docker"
			c.Args = append([]string{"docker", "exec", c.node.ExistingDockerContainer, name}, strArgs...)
		}
	}

	return c
}

func (c *Cmd) Start() error {
	// Add some IPC pipes to capture decryption secrets
	for envName, secretsType := range map[string]uint32{
		"SSLKEYLOGFILE": pcapgo.DSB_SECRETS_TYPE_TLS,
		"WG_KEYLOGFILE": pcapgo.DSB_SECRETS_TYPE_WIREGUARD,
	} {
		if pipe, err := c.node.network.KeyLogPipe(secretsType); err != nil {
			return fmt.Errorf("failed to open key log pipe: %w", err)
		} else if pipe != nil {
			c.extraEnvFile(envName, pipe)
		}
	}

	// Add tracing pipe
	if t := c.tracer(); t != nil {
		if pipe, err := t.Pipe(); err != nil {
			return fmt.Errorf("failed to create tracing pipe: %w", err)
		} else if pipe != nil {
			c.extraEnvFile("GONT_TRACEFILE", pipe)
		}
	}

	// Redirect process stdout/stderr to zapio.Writer
	var updateLogger func(*zap.Logger)
	if c.RedirectToLog || c.node.RedirectToLog || c.node.network.RedirectToLog {
		updateLogger = c.redirectToLog()
	}

	if len(c.StdoutWriters) > 0 {
		c.Stdout = io.MultiWriter(c.StdoutWriters...)
	}

	if len(c.StderrWriters) > 0 {
		c.Stderr = io.MultiWriter(c.StderrWriters...)
	}

	if err := c.Cmd.Start(); err != nil {
		return err
	}

	logger := c.logger.With(
		zap.Int("pid", c.Process.Pid),
	)

	if updateLogger != nil {
		updateLogger(logger)
	}

	logger.Info("Process started")

	return nil
}

func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		c.logger.Info("Failed to start", zap.Error(err))
		return err
	}
	err := c.Wait()

	logger := c.logger.With(
		zap.Int("pid", c.Process.Pid),
		zap.Int("rc", c.ProcessState.ExitCode()),
		zap.Duration("sys_time", c.ProcessState.SystemTime()),
	)

	if c.ProcessState.Success() {
		logger.Info("Process terminated successfully")
	} else {
		logger.Error("Process terminated with error code")
	}

	return err
}

// CombinedOutput runs the command and returns its combined standard
// output and standard error.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	var b bytes.Buffer

	c.StdoutWriters = append(c.StdoutWriters, &b)
	c.StderrWriters = append(c.StderrWriters, &b)

	err := c.Run()
	return b.Bytes(), err
}

func (c *Cmd) StdoutPipe() (io.ReadCloser, error) {
	rd, wr := io.Pipe()

	c.StdoutWriters = append(c.StdoutWriters, wr)

	return rd, nil
}

func (c *Cmd) StderrPipe() (io.ReadCloser, error) {
	rd, wr := io.Pipe()

	c.StderrWriters = append(c.StderrWriters, wr)

	return rd, nil
}

func (c *Cmd) tracer() *Tracer {
	if t := c.Tracer; t != nil {
		return t
	} else if t := c.node.network.Tracer; t != nil {
		return t
	} else {
		return nil
	}
}

func (c *Cmd) extraEnvFile(envName string, f *os.File) {
	fd := len(c.ExtraFiles) + 3
	c.ExtraFiles = append(c.ExtraFiles, f)
	c.Env = append(c.Env, fmt.Sprintf("%s=/proc/self/fd/%d", envName, fd))
}

func (c *Cmd) redirectToLog() func(*zap.Logger) {
	stdoutLog := &zapio.Writer{
		Log:   c.logger,
		Level: zap.InfoLevel,
	}

	stderrLog := &zapio.Writer{
		Log:   c.logger,
		Level: zap.WarnLevel,
	}

	c.StdoutWriters = append(c.StdoutWriters, stdoutLog)
	c.StderrWriters = append(c.StderrWriters, stderrLog)

	return func(l *zap.Logger) {
		stdoutLog.Log = l
		stderrLog.Log = l
	}
}

func stringifyArg(arg any) (string, bool) {
	switch arg := arg.(type) {
	case Node:
		return arg.Name(), true
	case fmt.Stringer:
		return arg.String(), true
	case string:
		return arg, true
	case int:
		return strconv.FormatInt(int64(arg), 10), true
	case uint:
		return strconv.FormatUint(uint64(arg), 10), true
	case int32:
		return strconv.FormatInt(int64(arg), 10), true
	case uint32:
		return strconv.FormatUint(uint64(arg), 10), true
	case int64:
		return strconv.FormatInt(arg, 10), true
	case uint64:
		return strconv.FormatUint(arg, 10), true
	case float32:
		return strconv.FormatFloat(float64(arg), 'f', -1, 32), true
	case float64:
		return strconv.FormatFloat(arg, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(arg), true
	default:
		return "", false
	}
}
