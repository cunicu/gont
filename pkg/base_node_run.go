// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"
	"io"
	"math/rand"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

func (n *BaseNode) Command(name string, args ...any) *Cmd {
	return command(name, n, args...)
}

func (n *BaseNode) Run(cmd string, args ...any) ([]byte, *exec.Cmd, error) {
	stdout, stderr, c, err := n.Start(cmd, args...)
	if err != nil {
		return nil, nil, err
	}

	combined := io.MultiReader(stdout, stderr)
	buf, err := io.ReadAll(combined)
	if err != nil {
		return nil, nil, err
	}

	if err = c.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, nil, err
		}
	}

	logger := n.logger.With(
		zap.Any("node", n),
		zap.String("cmd", cmd),
		zap.Any("cmd_args", args),
		zap.Int("pid", c.Process.Pid),
		zap.Int("rc", c.ProcessState.ExitCode()),
		zap.Duration("sys_time", c.ProcessState.SystemTime()),
	)

	if c.ProcessState.Success() {
		logger.Info("Process terminated successfully")
	} else {
		logger.Error("Process terminated with error code")
	}

	return buf, c, err
}

func (n *BaseNode) Start(cmd string, args ...any) (io.Reader, io.Reader, *exec.Cmd, error) {
	var err error
	var stdout, stderr io.Reader

	c := n.Command(cmd, args...)

	if stdout, err = c.StdoutPipe(); err != nil {
		return nil, nil, nil, err
	}

	if stderr, err = c.StderrPipe(); err != nil {
		return nil, nil, nil, err
	}

	logger := n.logger.With(
		zap.String("cmd", cmd),
		zap.Any("cmd_args", c.Args),
	)

	if err = c.Start(); err != nil {
		logger.Error("Failed to start", zap.Error(err))

		return nil, nil, c.Cmd, err
	}

	logger = logger.With(
		zap.Int("pid", c.Process.Pid),
	)

	logger.Info("Process started")

	if n.LogToDebug {
		slogger := zap.L().With(zap.Int("pid", c.Process.Pid))

		logStdout := &zapio.Writer{
			Log:   slogger,
			Level: zap.InfoLevel,
		}

		logStderr := &zapio.Writer{
			Log:   slogger,
			Level: zap.WarnLevel,
		}

		outReader, outWriter := io.Pipe()
		errReader, errWriter := io.Pipe()

		stdout = io.TeeReader(stdout, outWriter)
		stderr = io.TeeReader(stderr, errWriter)

		go io.Copy(logStdout, outReader)
		go io.Copy(logStderr, errReader)
	}

	return stdout, stderr, c.Cmd, nil
}

func (n *BaseNode) StartGo(script string, args ...any) (io.Reader, io.Reader, *exec.Cmd, error) {
	tmp := filepath.Join(n.network.VarPath, fmt.Sprintf("go-build-%d", rand.Intn(1<<16)))

	if out, _, err := n.network.HostNode.Run("go", "build", "-o", tmp, script); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to compile Go code: %w\n%s", err, string(out))
	}

	return n.Start(tmp, args...)
}

func (n *BaseNode) RunGo(script string, args ...any) ([]byte, *exec.Cmd, error) {
	tmp := filepath.Join(n.network.TmpPath, fmt.Sprintf("go-build-%d", rand.Intn(1<<16)))

	if _, _, err := n.network.HostNode.Run("go", "build", "-o", tmp, script); err != nil {
		return nil, nil, fmt.Errorf("failed to compile Go code: %w", err)
	}

	return n.Run(tmp, args...)
}
