// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"fmt"
	"io"
	"testing"

	g "github.com/stv0g/gont/pkg"
	"github.com/vishvananda/netns"
)

func prepare(t *testing.T) (*g.Network, *g.BaseNode) {
	n, err := g.NewNetwork(*nname, globalNetworkOptions...)
	if err != nil {
		t.Fatalf("Failed to create new network: %s", err)
	}

	n1, err := n.AddNode("n1")
	if err != nil {
		t.Fatalf("Failed to create node: %s", err)
	}

	return n, n1
}

func TestRun(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	// Run
	out, _, err := n1.Run("ip", "netns", "identify")
	if err != nil {
		t.Errorf("Failed to run identify: %s", err)
	}

	if string(out) != n1.Namespace.Name+"\n" {
		t.Errorf("Got invalid namespace: %s", string(out))
	}
}

func TestRunFunc(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	// Run
	if err := n1.RunFunc(func() error {
		handle, err := netns.Get()
		if err != nil {
			return err
		}

		if !handle.Equal(n1.NsHandle) {
			t.Fatalf("mismatching netns handles")
		}

		return nil
	}); err != nil {
		t.Errorf("Failed to run identify: %s", err)
	}
}

func TestRunGo(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	out, cmd, err := n1.RunGo("../cmd/gontc/gontc.go", "identify")
	if err != nil {
		t.Fatalf("Failed to run Go script: %s", err)
	}

	if !cmd.ProcessState.Exited() || !cmd.ProcessState.Success() {
		t.FailNow()
	}

	if string(out) != fmt.Sprintf("%s\n", n1) {
		t.FailNow()
	}
}

func TestEnter(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	exit, err := n1.Enter()
	if err != nil {
		t.FailNow()
	}
	defer exit()

	handle, err := netns.Get()
	if err != nil {
		t.FailNow()
	}

	if !handle.Equal(n1.NsHandle) {
		t.Fatalf("mismatching netns handles")
	}
}

func TestRunSimple(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	if _, _, err := n1.Run("true"); err != nil {
		t.Fail()
	}

	if _, _, err := n1.Run("false"); err == nil {
		t.Fail()
	}
}

func TestStart(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	stdout, _, cmd, err := n1.Start("ip", "netns", "identify")
	if err != nil {
		t.Errorf("Failed to run identify: %s", err)
	}

	var out []byte
	if out, err = io.ReadAll(stdout); err != nil {
		t.Errorf("Failed to read all: %s", err)
	}

	cmd.Wait()

	if !cmd.ProcessState.Exited() || !cmd.ProcessState.Success() {
		t.FailNow()
	}

	if string(out) != n1.Namespace.Name+"\n" {
		t.Errorf("Got invalid namespace: %s", string(out))
	}
}
