// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"bytes"
	"testing"

	g "cunicu.li/gont/v2/pkg"
	co "cunicu.li/gont/v2/pkg/options/cmd"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netns"
)

func prepare(t *testing.T) (*g.Network, *g.BaseNode) {
	n, err := g.NewNetwork(*nname)
	require.NoError(t, err, "Failed to create new network")

	n1, err := n.AddNode("n1")
	require.NoError(t, err, "Failed to create node")

	return n, n1
}

func TestRun(t *testing.T) {
	n, n1 := prepare(t)
	defer n.MustClose()

	// Run
	outp := &bytes.Buffer{}
	_, err := n1.Run("ip", "netns", "identify",
		co.Stdout(outp))
	require.NoError(t, err, "Failed to run identify")
	require.Equal(t, outp.String(), n1.Namespace.Name+"\n", "Got invalid namespace")
}

func TestRunFunc(t *testing.T) {
	n, n1 := prepare(t)
	defer n.MustClose()

	// Run
	err := n1.RunFunc(func() error {
		handle, err := netns.Get()
		if err != nil {
			return err
		}

		require.True(t, handle.Equal(n1.NsHandle), "Mismatching netns handles")

		return nil
	})
	require.NoError(t, err, "Failed to run identify")
}

func TestRunGo(t *testing.T) {
	n, n1 := prepare(t)
	defer n.MustClose()

	outp := &bytes.Buffer{}
	cmd, err := n1.RunGo("../cmd/gontc", "identify",
		co.Stdout(outp))
	require.NoError(t, err, "Failed to run Go script")

	require.True(t, cmd.ProcessState.Exited(), "Process did not exit")
	require.True(t, cmd.ProcessState.Success(), "Process did not succeed")
	require.Equal(t, outp.String(), n1.String()+"\n")
}

func TestEnter(t *testing.T) {
	n, n1 := prepare(t)
	defer n.MustClose()

	exit, err := n1.Enter()
	require.NoError(t, err, "Failed to enter namespace")
	defer exit()

	handle, err := netns.Get()
	require.NoError(t, err, "Failed to get netns handle")
	require.True(t, handle.Equal(n1.NsHandle), "Mismatching netns handles")
}

func TestRunSimple(t *testing.T) {
	n, n1 := prepare(t)
	defer n.MustClose()

	_, err := n1.Run("true")
	require.NoError(t, err)

	_, err = n1.Run("false")
	require.Error(t, err)
}

func TestStart(t *testing.T) {
	n, n1 := prepare(t)
	defer n.MustClose()

	outp := &bytes.Buffer{}
	cmd, err := n1.Start("ip", "netns", "identify",
		co.Stdout(outp))
	require.NoError(t, err, "Failed to run identify")

	err = cmd.Wait()
	require.NoError(t, err, "Failed to wait")

	require.True(t, cmd.ProcessState.Exited(), "Process did not exit")
	require.True(t, cmd.ProcessState.Success(), "Process did not succeed")

	require.Equal(t, outp.String(), n1.Namespace.Name+"\n", "Got invalid namespace")
}
