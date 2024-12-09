// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"fmt"
	"testing"
	"time"

	g "cunicu.li/gont/v2/pkg"
	sdo "cunicu.li/gont/v2/pkg/options/systemd"
	"github.com/stretchr/testify/require"
)

func TestCGroup(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h, err := n.AddHost("h1")
	require.NoError(t, err)

	cmd := h.Command("cat", "/proc/self/cgroup")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)

	expectedCgroup := fmt.Sprintf("0::/gont.slice/gont-%s.slice/gont-%s-%s.slice/gont-run-%d.scope\n", n.Name, n.Name, h.Name(), cmd.ProcessState.Pid())
	require.Equal(t, expectedCgroup, string(out))
}

func TestCGroupPropertyNetwork(t *testing.T) {
	n, err := g.NewNetwork("", sdo.MemoryMax(5<<20))
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h, err := n.AddHost("h1")
	require.NoError(t, err)

	cmd := h.Command("bash", "-c", "systemctl show "+n.Unit()+" | grep ^MemoryMax=")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)

	outExpected := fmt.Sprintf("MemoryMax=%d\n", 5<<20)
	require.Equal(t, outExpected, string(out))
}

func TestCGroupPropertyHost(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h, err := n.AddHost("h1", sdo.MemoryMax(5<<20))
	require.NoError(t, err)

	cmd := h.Command("bash", "-c", "systemctl show "+h.Unit()+" | grep ^MemoryMax=")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)

	outExpected := fmt.Sprintf("MemoryMax=%d\n", 5<<20)
	require.Equal(t, outExpected, string(out))
}

func TestCGroupPropertyCommand(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h, err := n.AddHost("h1")
	require.NoError(t, err)

	cmd := h.Command("bash", "-c", "systemctl show gont-run-$$.scope | grep ^MemoryMax=", sdo.MemoryMax(5<<20))
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)

	outExpected := fmt.Sprintf("MemoryMax=%d\n", 5<<20)
	require.Equal(t, outExpected, string(out))
}

func TestCGroupTeardown(t *testing.T) {
	n, err := g.NewNetwork("")
	require.NoError(t, err, "Failed to create network")

	h, err := n.AddHost("h1", sdo.MemoryMax(5<<20))
	require.NoError(t, err)

	cmd := h.Command("sleep", 3600)
	err = cmd.Start()
	require.NoError(t, err)

	exited := make(chan bool)
	go func() {
		cmd.Wait() //nolint:errcheck
		close(exited)
	}()

	time.Sleep(10 * time.Millisecond)

	err = n.Close()
	require.NoError(t, err)

	select {
	case <-exited:

	case <-time.After(10 * time.Millisecond):
		require.Fail(t, "Process did not terminate")
	}
}
