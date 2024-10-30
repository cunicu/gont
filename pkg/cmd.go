// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	sdbus "github.com/coreos/go-systemd/v22/dbus"
	"github.com/godbus/dbus/v5"
	"github.com/gopacket/gopacket/pcapgo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

//nolint:gochecknoglobals
var DefaultPreserveEnvVars = []string{
	"PATH",
}

type ExecCmdOption interface {
	ApplyExecCmd(*exec.Cmd)
}

type CmdOption interface {
	ApplyCmd(*Cmd)
}

type Cmd struct {
	*CGroup
	*exec.Cmd

	// Options
	Tracer          *Tracer
	Debugger        *Debugger
	RedirectToLog   bool
	DisableASLR     bool
	Context         context.Context
	PreserveEnvVars []string

	StdoutWriters []io.Writer
	StderrWriters []io.Writer

	debuggerInstance *debuggerInstance
	node             *BaseNode
	logger           *zap.Logger
}

func (n *BaseNode) Command(name string, args ...any) *Cmd {
	c := &Cmd{
		node: n,
	}

	c.CGroup, _ = NewCGroup(c.node.sdConn, "scope", "")

	strArgs := []string{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case ExecCmdOption:
		case CmdOption:
			arg.ApplyCmd(c)
		case CGroupOption:
			arg.ApplyCGroup(c.CGroup)
		default:
			if strArg, ok := stringifyArg(arg); ok {
				strArgs = append(strArgs, strArg)
			}
		}
	}

	if c.Cmd == nil {
		if c.Context != nil {
			c.Cmd = exec.CommandContext(c.Context, name, strArgs...)
		} else {
			c.Cmd = exec.Command(name, strArgs...)
		}
	} else {
		c.Cmd.Args = append(c.Cmd.Args, strArgs...)
	}

	for _, arg := range args {
		switch arg := arg.(type) {
		case ExecCmdOption:
			arg.ApplyExecCmd(c.Cmd)
		}
	}

	c.logger = n.logger.Named("cmd").With(
		zap.String("path", name),
		zap.Strings("args", strArgs),
	)

	// Preserve some environment variables from the parent process
	if c.PreserveEnvVars == nil {
		c.PreserveEnvVars = DefaultPreserveEnvVars
	}

	passEnv := func(name string) {
		if value := os.Getenv(name); value != "" {
			c.Env = append(c.Env, fmt.Sprintf("%s=%s", name, value))
		}
	}

	for _, name := range c.PreserveEnvVars {
		passEnv(name)
	}

	// Actual namespace switching is done similar to Docker's reexec
	// in a forked version of ourself by passing all required details
	// in environment variables.
	if !c.node.isHostNode {
		if c.node.ExistingDockerContainer == "" {
			c.Path = "/proc/self/exe"
			c.Env = append(c.Env,
				"GONT_UNSHARE=true",
				"GONT_NODE="+c.node.name,
				"GONT_NETWORK="+c.node.network.Name)

			passEnv("GONT_SKIP_MISSING_MOUNTPOINT")
		} else {
			c.Path = "/usr/bin/docker"
			c.Args = append([]string{"docker", "exec", c.node.ExistingDockerContainer, name}, strArgs...)
		}

		if c.DisableASLR {
			c.Env = append(c.Env,
				"GONT_DISABLE_ASLR=true",
			)
		}
	}

	return c
}

func (c *Cmd) Start() (err error) {
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

	// We need to start the process in a stopped state for two reasons
	// 1. Attaching the Delve debugger before execution
	//    commences in order to allow for breakpoints early
	//    in the execution.
	// 2. Attaching the process into the new Systemd Scope Unit
	if err := c.stoppedStart(); err != nil {
		return err
	}

	if d := c.debugger(); d != nil {
		if c.debuggerInstance, err = d.start(c.Cmd); err != nil {
			return err
		}
	}

	// Add PID as field to logger after the process has been started
	if updateLogger != nil {
		updateLogger(c.logger.With(
			zap.Int("pid", c.Process.Pid),
		))
	}

	// Start CGroup scope and attach process to it
	c.CGroup.Name = fmt.Sprintf("gont-run-%d", c.Process.Pid)
	c.CGroup.Properties = append(c.CGroup.Properties,
		sdbus.Property{
			Name:  "Slice",
			Value: dbus.MakeVariant(c.node.Unit()),
		},
		sdbus.PropPids(uint32(c.Process.Pid)), //nolint:gosec
	)

	if err := c.CGroup.Start(); err != nil {
		return fmt.Errorf("failed to start CGroup scope: %w", err)
	}

	// Signal child that that it is ready to proceed
	if err := c.Process.Signal(syscall.SIGCONT); err != nil {
		return fmt.Errorf("failed to send continuation signal to child: %w", err)
	}

	return nil
}

func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}

	return c.Wait()
}

func (c *Cmd) Wait() error {
	if d := c.debuggerInstance; d != nil {
		<-d.stop
	}

	return c.Cmd.Wait()
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
	} else if t := c.node.Tracer; t != nil {
		return t
	} else if t := c.node.network.Tracer; t != nil {
		return t
	} else {
		return nil
	}
}

func (c *Cmd) debugger() *Debugger {
	if d := c.Debugger; d != nil {
		return d
	} else if d := c.node.Debugger; d != nil {
		return d
	} else if d := c.node.network.Debugger; d != nil {
		return d
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

// stoppedStart starts the child in a stopped state (SIGSTOP).
//
// Starting the process in a stopped state requires the following sequence:
//
//  1. We start ourself as a traced sub-process:
//     Cmd.Path = "/proc/self/exe"
//     Cmd.Args = <args of new child>
//     Cmd.SysProcAttr.Ptrace = true
//
//  2. In response to 1), the child process calls immediately ptrace(PTRACE_TRACEME, ...)
//
//  3. The child now becomes a tracee and enters a Ptrace-stop
//
//  4. The parent waits for the tracee to enter the Ptrace-stop
//     wait4(pid, &ws, 0, NULL)
//     WSTOPSIG(ws) == SIGTRAP
//
//  5. We now enable the Ptrace-exec-stop:
//     ptrace(PTRACE_SETOPTIONS, pid, 0, PTRACE_O_TRACEEXEC)
//
//  6. The tracee execution is continued and uses execve() to start the actual child process
//
//  7. The tracee enters the Ptrace exec-stop
//
//  8. While stopped, we send a SIGSTOP to the tracee to provoke a Ptrace signal-delivery-stop.
//
//  9. We continue the execution of the tracee until the Ptrace signal-delivery-stop.
//
//  9. We detach from the tracee and inject a SIGSTOP signal
//     ptrace(PTRACE_DETACH, pid, 0, SIGSTOP)
//
// 10) The parent process has detached from the tracee and the tracee is stopped due to the injected SIGSTOP in 9)
func (c *Cmd) stoppedStart() error {
	if c.Cmd.SysProcAttr == nil {
		c.Cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	c.SysProcAttr.Setpgid = true
	c.SysProcAttr.Ptrace = true

	if err := c.Cmd.Start(); err != nil {
		return err
	}

	pgid, err := syscall.Getpgid(c.Process.Pid)
	if err != nil {
		return fmt.Errorf("failed to get pgid: %w", err)
	}

	for {
		var ws syscall.WaitStatus
		wpid, err := syscall.Wait4(-pgid, &ws, syscall.WALL, nil)
		if err != nil {
			return err
		}

		// c.logger.Debug("Stopped",
		// 	zap.String("signal", ws.Signal().String()),
		// 	zap.String("stop_signal", ws.StopSignal().String()),
		// 	zap.Int("trap_cause", ws.TrapCause()))

		if ws.Exited() {
			return fmt.Errorf("process exited prematurely")
		}

		if !ws.Stopped() {
			continue
		}

		switch ws.StopSignal() {
		case syscall.SIGTRAP:
			if c.node.isHostNode || ws.TrapCause() == syscall.PTRACE_EVENT_EXEC {
				if err := syscall.Tgkill(c.Process.Pid, wpid, syscall.SIGSTOP); err != nil {
					return err
				}
			} else if ws.TrapCause() == 0 {
				if err := syscall.PtraceSetOptions(wpid, syscall.PTRACE_O_TRACEEXEC); err != nil {
					return err
				}
			}

		case syscall.SIGSTOP:
			if err := ptrace(syscall.PTRACE_DETACH, wpid, 0, uintptr(syscall.SIGSTOP)); err != nil {
				return err
			}

			c.logger.Debug("Detached from tracee")

			return nil
		}

		if err = syscall.PtraceCont(wpid, 0); err != nil {
			return err
		}
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
