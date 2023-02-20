// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"bytes"
	"fmt"
	"testing"

	g "github.com/stv0g/gont/pkg"
	co "github.com/stv0g/gont/pkg/options/cmd"
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
	outp := &bytes.Buffer{}
	if _, err := n1.Run("ip", "netns", "identify",
		co.Stdout(outp),
	); err != nil {
		t.Errorf("Failed to run identify: %s", err)
	}

	if outp.String() != n1.Namespace.Name+"\n" {
		t.Errorf("Got invalid namespace: %s", outp.String())
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

	outp := &bytes.Buffer{}
	cmd, err := n1.RunGo("../cmd/gontc/gontc.go", "identify",
		co.Stdout(outp),
	)
	if err != nil {
		t.Fatalf("Failed to run Go script: %s", err)
	}

	if !cmd.ProcessState.Exited() || !cmd.ProcessState.Success() {
		t.FailNow()
	}

	if outp.String() != fmt.Sprintf("%s\n", n1) {
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

	if _, err := n1.Run("true"); err != nil {
		t.Fail()
	}

	if _, err := n1.Run("false"); err == nil {
		t.Fail()
	}
}

func TestStart(t *testing.T) {
	n, n1 := prepare(t)
	defer n.Close()

	outp := &bytes.Buffer{}
	cmd, err := n1.Start("ip", "netns", "identify",
		co.Stdout(outp),
	)
	if err != nil {
		t.Errorf("Failed to run identify: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}

	if !cmd.ProcessState.Exited() || !cmd.ProcessState.Success() {
		t.FailNow()
	}

	if outp.String() != n1.Namespace.Name+"\n" {
		t.Errorf("Got invalid namespace: %s", outp.String())
	}
}
