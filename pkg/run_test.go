// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	g "github.com/stv0g/gont/pkg"
	co "github.com/stv0g/gont/pkg/options/cmd"
	"github.com/vishvananda/netns"
)

func prepare(t *testing.T) (*g.Network, *g.BaseNode) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	assert.NoError(t, err, "Failed to create new network")

	n1, err := n.AddNode("n1")
	assert.NoError(t, err, "Failed to create node")

	return n, n1
}

func TestRun(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	// Run
	outp := &bytes.Buffer{}
	_, err := n1.Run("ip", "netns", "identify",
		co.Stdout(outp))
	assert.NoError(t, err, "Failed to run identify")
	assert.Equal(t, outp.String(), n1.Namespace.Name+"\n", "Got invalid namespace")
}

func TestRunFunc(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	// Run
	err := n1.RunFunc(func() error {
		handle, err := netns.Get()
		if err != nil {
			return err
		}

		assert.True(t, handle.Equal(n1.NsHandle), "Mismatching netns handles")

		return nil
	})
	assert.NoError(t, err, "Failed to run identify")
}

func TestRunGo(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	outp := &bytes.Buffer{}
	cmd, err := n1.RunGo("../cmd/gontc/gontc.go", "identify",
		co.Stdout(outp))
	assert.NoError(t, err, "Failed to run Go script")

	assert.True(t, cmd.ProcessState.Exited(), "Process did not exit")
	assert.True(t, cmd.ProcessState.Success(), "Process did not succeed")
	assert.Equal(t, outp.String(), n1.String()+"\n")
}

func TestEnter(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	exit, err := n1.Enter()
	assert.NoError(t, err, "Failed to enter namespace")
	defer exit()

	handle, err := netns.Get()
	assert.NoError(t, err, "Failed to get netns handle")
	assert.True(t, handle.Equal(n1.NsHandle), "Mismatching netns handles")
}

func TestRunSimple(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	_, err := n1.Run("true")
	assert.NoError(t, err)

	_, err = n1.Run("false")
	assert.Error(t, err)
}

func TestStart(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	outp := &bytes.Buffer{}
	cmd, err := n1.Start("ip", "netns", "identify",
		co.Stdout(outp))
	assert.NoError(t, err, "Failed to run identify")

	err = cmd.Wait()
	assert.NoError(t, err, "Failed to wait")

	assert.True(t, cmd.ProcessState.Exited(), "Process did not exit")
	assert.True(t, cmd.ProcessState.Success(), "Process did not succeed")

	assert.Equal(t, outp.String(), n1.Namespace.Name+"\n", "Got invalid namespace")
}
