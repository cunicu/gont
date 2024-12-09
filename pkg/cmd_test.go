// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	co "cunicu.li/gont/v2/pkg/options/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdStdinStdout(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	hn, err := n.AddHost("host", o.HostNamespace)
	require.NoError(t, err, "Failed to create host node")

	stdin := bytes.NewBuffer([]byte("Hello World"))
	stdout := bytes.NewBuffer(nil)

	_, err = hn.Run("cat", co.Stdin(stdin), co.Stdout(stdout))
	require.NoError(t, err, "Failed to run")
	require.Equal(t, "Hello World", stdout.String())
}

func TestCmdStdinStderr(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	hn, err := n.AddHost("host", o.HostNamespace)
	require.NoError(t, err, "Failed to create host node")

	stdin := bytes.NewBuffer([]byte("Hello World"))
	stderr := bytes.NewBuffer(nil)

	_, err = hn.Run("sh", "-c", "cat 1>&2", co.Stdin(stdin), co.Stderr(stderr))
	require.NoError(t, err, "Failed to run")
	require.Equal(t, "Hello World", stderr.String())
}

func TestCmdStdinCombined(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	hn, err := n.AddHost("host", o.HostNamespace)
	require.NoError(t, err, "Failed to create host node")

	stdin := bytes.NewBuffer([]byte("Hello World"))
	combined := bytes.NewBuffer(nil)

	_, err = hn.Run("cat", co.Stdin(stdin), co.Combined(combined))
	require.NoError(t, err, "Failed to run")
	require.Equal(t, "Hello World", combined.String())
}

func TestCmdArguments(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	hn, err := n.AddHost("host", o.HostNamespace)
	require.NoError(t, err, "Failed to create host node")

	combined := bytes.NewBuffer(nil)

	_, err = hn.Run("sh", "-c", co.Arg("echo -n \"Hello World\""), co.Combined(combined))
	require.NoError(t, err, "Failed to run")
	require.Equal(t, "Hello World", combined.String())

	combined.Reset()

	_, err = hn.Run("sh", co.Args("-c", "echo -n \"Hello World\""), co.Combined(combined))
	require.NoError(t, err, "Failed to run")
	require.Equal(t, "Hello World", combined.String())
}

func TestCmdDir(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	hn, err := n.AddHost("host", o.HostNamespace)
	require.NoError(t, err, "Failed to create host node")

	combined := bytes.NewBuffer(nil)

	_, err = hn.Run("pwd", co.Combined(combined), co.Dir("/var/lib"))
	require.NoError(t, err, "Failed to run")
	require.Equal(t, "/var/lib\n", combined.String())
}

func TestCmdEnv(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	hn, err := n.AddHost("host", o.HostNamespace)
	require.NoError(t, err, "Failed to create host node")

	combined := bytes.NewBuffer(nil)

	_, err = hn.Run("sh", "-c", "echo -n ${MYKEY}", co.Combined(combined), co.EnvVar("MYKEY", "MYVALUE"))
	require.NoError(t, err, "Failed to run")
	require.Equal(t, "MYVALUE", combined.String())

	combined.Reset()

	_, err = hn.Run("sh", "-c", "echo -n ${MYKEY}", co.Combined(combined), co.Env("MYKEY=MYVALUE"))
	require.NoError(t, err, "Failed to run")
	require.Equal(t, "MYVALUE", combined.String())

	combined.Reset()

	_, err = hn.Run("sh", "-c", "echo -n ${MYKEY}", co.Combined(combined), co.Envs([]string{"MYKEY=MYVALUE"}))
	require.NoError(t, err, "Failed to run")
	require.Equal(t, "MYVALUE", combined.String())
}

func TestCmdExtraFile(t *testing.T) {
	t.Skip()

	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	hn, err := n.AddHost("host", o.HostNamespace)
	require.NoError(t, err, "Failed to create host node")

	combined := bytes.NewBuffer(nil)

	rd, wr, err := os.Pipe()
	require.NoError(t, err)

	c, err := hn.Start("cat", "/proc/self/fd/3", (*co.ExtraFile)(rd), co.Combined(combined))
	require.NoError(t, err)

	_, err = wr.Write([]byte("Hello World"))
	require.NoError(t, err)

	err = wr.Close()
	require.NoError(t, err)

	err = c.Wait()
	require.NoError(t, err)

	require.Equal(t, "Hello World", combined.String())
}

func TestCmdContext(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	hn, err := n.AddHost("host", o.HostNamespace)
	require.NoError(t, err, "Failed to create host node")

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	c, err := hn.Run("sleep", 100, co.Context{Context: ctx})
	require.NotNil(t, err)
	ws, ok := c.ProcessState.Sys().(syscall.WaitStatus)
	require.True(t, ok)
	require.True(t, ws.Signaled())
	require.Equal(t, syscall.SIGKILL, ws.Signal())
}

func TestIProute2Files(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	beep, err := n.AddHost("beep")
	require.NoError(t, err)

	cmd := beep.Command("ip", "addr")
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err)

	t.Logf("Output: %s", out)
}

func TestExecCmd(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	hn, err := n.AddHost("host", o.HostNamespace)
	require.NoError(t, err, "Failed to create host node")

	stdin := bytes.NewBuffer([]byte("Hello World"))
	stdout := bytes.NewBuffer(nil)

	cmd := exec.Command("cat")
	cmd.Stdin = stdin
	cmd.Stdout = stdout

	_, err = hn.Run("is-ignored", co.Command(cmd))
	require.NoError(t, err, "Failed to run")
	require.Equal(t, "Hello World", stdout.String())
}
