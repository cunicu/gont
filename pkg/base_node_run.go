package gont

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func (n *BaseNode) Command(name string, arg ...string) *exec.Cmd {
	// Actual namespace switching is done similar to Docker's reexec
	// in a forked version of ourself by passing all required details
	// in environment variables.

	c := exec.Command(name, arg...)

	if !n.NsHandle.Equal(n.network.HostNode.NsHandle) {
		if n.ExistingDockerContainer == "" {
			c.Path = "/proc/self/exe"
			c.Env = append(os.Environ(),
				"GONT_UNSHARE=exec",
				"GONT_NODE="+n.name,
				"GONT_NETWORK="+n.network.Name)
		} else {
			c.Path = "/usr/bin/docker"
			c.Args = append([]string{"docker", "exec", n.ExistingDockerContainer, name}, arg...)
		}
	}

	return c
}

func (n *BaseNode) Run(cmd string, args ...interface{}) ([]byte, *exec.Cmd, error) {
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

	rlogger := log.WithFields(log.Fields{
		"node":     n,
		"cmd":      cmd,
		"cmd_args": args,
		"pid":      c.Process.Pid,
		"rc":       c.ProcessState.ExitCode(),
		"sys_time": c.ProcessState.SystemTime(),
	})

	f := rlogger.Info
	if !c.ProcessState.Success() {
		f = rlogger.Error
	}
	f("Process terminated")

	return buf, c, err
}

func (n *BaseNode) Start(cmd string, args ...interface{}) (io.Reader, io.Reader, *exec.Cmd, error) {
	var err error
	var stdout, stderr io.Reader

	strargs := []string{}
	for _, arg := range args {
		var strarg string
		switch arg := arg.(type) {
		case Node:
			strarg = arg.Name()
		case fmt.Stringer:
			strarg = arg.String()
		case string:
			strarg = arg
		case int:
			strarg = strconv.FormatInt(int64(arg), 10)
		case uint:
			strarg = strconv.FormatUint(uint64(arg), 10)
		case int32:
			strarg = strconv.FormatInt(int64(arg), 10)
		case uint32:
			strarg = strconv.FormatUint(uint64(arg), 10)
		case int64:
			strarg = strconv.FormatInt(arg, 10)
		case uint64:
			strarg = strconv.FormatUint(arg, 10)
		case float32:
			strarg = strconv.FormatFloat(float64(arg), 'f', -1, 32)
		case float64:
			strarg = strconv.FormatFloat(arg, 'f', -1, 64)
		case bool:
			strarg = strconv.FormatBool(arg)
		default:
			return nil, nil, nil, fmt.Errorf("invalid argument: %v", arg)
		}

		strargs = append(strargs, strarg)
	}

	c := n.Command(cmd, strargs...)

	if stdout, err = c.StdoutPipe(); err != nil {
		return nil, nil, nil, err
	}

	if stderr, err = c.StderrPipe(); err != nil {
		return nil, nil, nil, err
	}

	logger := log.WithFields(log.Fields{
		"node":     n,
		"cmd":      cmd,
		"cmd_args": strargs,
	})

	if err = c.Start(); err != nil {
		logger.WithError(err).Error("Failed to start")

		return nil, nil, c, err
	}

	logger = logger.WithField("pid", c.Process.Pid)

	logger.Info("Process started")

	if log.GetLevel() >= log.DebugLevel {
		slogger := log.WithFields(log.Fields{
			"pid": c.Process.Pid,
		})

		outReader, outWriter := io.Pipe()
		errReader, errWriter := io.Pipe()

		stdout = io.TeeReader(stdout, outWriter)
		stderr = io.TeeReader(stderr, errWriter)

		go io.Copy(slogger.WriterLevel(log.InfoLevel), outReader)
		go io.Copy(slogger.WriterLevel(log.WarnLevel), errReader)
	}

	return stdout, stderr, c, nil
}

func (n *BaseNode) StartGo(script string, arg ...interface{}) (io.Reader, io.Reader, *exec.Cmd, error) {
	tmp := filepath.Join(n.network.BasePath, fmt.Sprintf("go-build-%d", rand.Intn(1<<16)))
	_, _, err := n.network.HostNode.Run("go", "build", "-o", tmp, script)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to compile Go code: %w", err)
	}

	return n.Start(tmp, arg...)
}

func (n *BaseNode) RunGo(script string, arg ...interface{}) ([]byte, *exec.Cmd, error) {
	tmp := filepath.Join(n.network.BasePath, fmt.Sprintf("go-build-%d", rand.Intn(1<<16)))
	_, _, err := n.network.HostNode.Run("go", "build", "-o", tmp, script)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compile Go code: %w", err)
	}

	return n.Run(tmp, arg...)
}
