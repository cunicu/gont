// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"

	unixx "cunicu.li/gont/v2/internal/unix"
	sdbus "github.com/coreos/go-systemd/v22/dbus"
	"github.com/godbus/dbus/v5"
	"github.com/gopacket/gopacket/pcapgo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
	"golang.org/x/sys/unix"
)

var errProcessExitedPrematurely = errors.New("process exited prematurely")

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
	Slice           string
	Scope           string
	CGroupOptions   []Option

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

	strArgs := []string{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case ExecCmdOption:
		case CmdOption:
			arg.ApplyCmd(c)
		case CGroupOption:
			c.CGroupOptions = append(c.CGroupOptions, arg)
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
		if arg, ok := arg.(ExecCmdOption); ok {
			arg.ApplyExecCmd(c.Cmd)
		}
	}

	c.logger = n.logger.Named("cmd").With(
		zap.String("path", name),
		zap.Strings("args", strArgs),
	)

	setEnv := func(name, value string) {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", name, value))
	}

	passEnv := func(name string) {
		if value := os.Getenv(name); value != "" {
			setEnv(name, value)
		}
	}

	// Preserve some environment variables from the parent process
	if c.PreserveEnvVars == nil {
		c.PreserveEnvVars = DefaultPreserveEnvVars
	}

	for _, name := range c.PreserveEnvVars {
		passEnv(name)
	}

	// Actual namespace switching is done similar to Docker's reexec
	// in a forked version of ourself by passing all required details
	// in environment variables.
	if c.node.ExistingDockerContainer == "" {
		c.Path = "/proc/self/exe"

		setEnv("GONT_UNSHARE", "true")
		setEnv("GONT_NODE", c.node.name)
		setEnv("GONT_NETWORK", c.node.network.Name)
		passEnv("GONT_SKIP_MISSING_MOUNTPOINT")
	} else {
		c.Path = "/usr/bin/docker"
		c.Args = append([]string{"docker", "exec", c.node.ExistingDockerContainer, name}, strArgs...)
	}

	if c.DisableASLR {
		setEnv("GONT_DISABLE_ASLR", "true")
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
			if !errors.Is(err, errNoKeyLogs) {
				return fmt.Errorf("failed to open key log pipe: %w", err)
			}
		} else {
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

	// We need to start the process in a stopped state for two reasons:
	// 1. Attaching the Delve debugger before execution
	//    commences in order to allow for breakpoints early in the execution.
	// 2. Moving the process into the new Systemd Scope Unit / CGroup.
	var pid, pidfd int
	if pid, pidfd, err = c.stoppedStart(); err != nil {
		return err
	}

	if d := c.debugger(); d != nil {
		if c.debuggerInstance, err = d.newInstance(c.Cmd); err != nil {
			return err
		}
	}

	// Add PID as field to logger after the process has been started
	if updateLogger != nil {
		updateLogger(c.logger.With(
			zap.Int("pid", pid),
		))
	}

	if c.Scope == "" {
		c.Scope = fmt.Sprintf("gont-run-%d", pid)
	}

	// Start CGroup scope and attach process to it
	if c.CGroup, err = NewCGroup(c.node.sdConn, "scope", c.Scope, c.CGroupOptions...); err != nil {
		return fmt.Errorf("failed to create cgroup: %w", err)
	}

	c.CGroup.Properties = append(c.CGroup.Properties,
		sdbus.Property{
			Name:  "Slice",
			Value: dbus.MakeVariant(c.node.Unit()),
		},
	)

	if unixx.PidFDWorks() {
		c.CGroup.Properties = append(c.CGroup.Properties, sdbus.Property{
			Name:  "PIDFDs",
			Value: dbus.MakeVariant([]dbus.UnixFD{dbus.UnixFD(pidfd)}), //nolint:gosec
		})
	} else {
		c.CGroup.Properties = append(c.CGroup.Properties, sdbus.Property{
			Name:  "PIDs",
			Value: dbus.MakeVariant([]uint{uint(pidfd)}), //nolint:gosec
		})
	}

	if err := c.CGroup.Start(); err != nil {
		return fmt.Errorf("failed to start cgroup: %w", err)
	}

	// Signal child that that it is ready to proceed
	if err := c.Process.Signal(syscall.SIGCONT); err != nil {
		return fmt.Errorf("failed to send continuation signal to child: %w", err)
	}

	if di := c.debuggerInstance; di != nil {
		go di.run()
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
	}

	return nil
}

func (c *Cmd) debugger() *Debugger {
	if d := c.Debugger; d != nil {
		return d
	} else if d := c.node.Debugger; d != nil {
		return d
	} else if d := c.node.network.Debugger; d != nil {
		return d
	}

	return nil
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
func (c *Cmd) stoppedStart() (pid int, pidFD int, err error) {
	if c.Cmd.SysProcAttr == nil {
		c.Cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	c.Cmd.SysProcAttr.Setpgid = true
	c.Cmd.SysProcAttr.Ptrace = true
	c.Cmd.SysProcAttr.PidFD = &pidFD

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := c.Cmd.Start(); err != nil {
		return -1, -1, err
	}

	for {
		var si unixx.SiginfoChld
		if err := unixx.Waitid(unix.P_PIDFD, pidFD, &si, unix.WSTOPPED|unix.WEXITED, nil); err != nil {
			return -1, -1, fmt.Errorf("failed to wait for process: %w", err)
		}

		switch si.Code {
		case unixx.CLD_EXITED:
			return -1, -1, errProcessExitedPrematurely

		case unixx.CLD_TRAPPED:
			signal := syscall.Signal(si.Status & 0xff)
			trapCause := si.Status >> 8

			switch signal {
			case syscall.SIGTRAP:
				if trapCause == syscall.PTRACE_EVENT_EXEC {
					if err := syscall.Tgkill(c.Process.Pid, si.Pid, syscall.SIGSTOP); err != nil {
						return -1, -1, fmt.Errorf("failed to send SIGSTOP to child: %w", err)
					}
				} else {
					if err := syscall.PtraceSetOptions(si.Pid, syscall.PTRACE_O_TRACEEXEC); err != nil {
						return -1, -1, fmt.Errorf("failed to set pstrace options: %w", err)
					}
				}

			case syscall.SIGSTOP:
				if err := unixx.Ptrace(syscall.PTRACE_DETACH, si.Pid, 0, uintptr(syscall.SIGSTOP)); err != nil {
					return -1, -1, fmt.Errorf("failed to detach from tracee: %w", err)
				}

				return si.Pid, pidFD, nil

			default:
			}
		}

		if err = syscall.PtraceCont(si.Pid, 0); err != nil {
			return -1, -1, fmt.Errorf("failed to continue tracee: %w", err)
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
