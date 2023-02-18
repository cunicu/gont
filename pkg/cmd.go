// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/gopacket/gopacket/pcapgo"
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
	node   *BaseNode
	Tracer *Tracer
}

func command(name string, n *BaseNode, args ...any) *Cmd {
	c := &Cmd{
		node: n,
	}

	// Actual namespace switching is done similar to Docker's reexec
	// in a forked version of ourself by passing all required details
	// in environment variables.

	strArgs, nonStrArgs := stringifyArgs(args)

	c.Cmd = exec.Command(name, strArgs...)

	for _, arg := range nonStrArgs {
		switch arg := arg.(type) {
		case ExecCmdOption:
			arg.ApplyExecCmd(c.Cmd)
		case CmdOption:
			arg.ApplyCmd(c)
		}
	}

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

	return c.Cmd.Start()
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

func stringifyArgs(args []any) ([]string, []any) {
	strArgs := []string{}
	nonStrArgs := []any{}

	for _, arg := range args {
		switch arg := arg.(type) {
		case Node:
			strArgs = append(strArgs, arg.Name())
		case fmt.Stringer:
			strArgs = append(strArgs, arg.String())
		case string:
			strArgs = append(strArgs, arg)
		case int:
			strArgs = append(strArgs, strconv.FormatInt(int64(arg), 10))
		case uint:
			strArgs = append(strArgs, strconv.FormatUint(uint64(arg), 10))
		case int32:
			strArgs = append(strArgs, strconv.FormatInt(int64(arg), 10))
		case uint32:
			strArgs = append(strArgs, strconv.FormatUint(uint64(arg), 10))
		case int64:
			strArgs = append(strArgs, strconv.FormatInt(arg, 10))
		case uint64:
			strArgs = append(strArgs, strconv.FormatUint(arg, 10))
		case float32:
			strArgs = append(strArgs, strconv.FormatFloat(float64(arg), 'f', -1, 32))
		case float64:
			strArgs = append(strArgs, strconv.FormatFloat(arg, 'f', -1, 64))
		case bool:
			strArgs = append(strArgs, strconv.FormatBool(arg))
		default:
			nonStrArgs = append(nonStrArgs, arg)
		}
	}

	return strArgs, nonStrArgs
}
