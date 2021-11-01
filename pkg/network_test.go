package gont_test

import (
	"fmt"
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	"github.com/vishvananda/netns"
)

func hasNetwork(name string) bool {
	for _, n := range g.NetworkNames() {
		if n == name {
			return true
		}
	}
	return false
}

func TestNamedNetwork(t *testing.T) {
	var (
		err error
		n   *g.Network
	)

	name := g.GenerateNetworkName()
	ns := fmt.Sprintf("gont-%s-h1", name)

	if n, err = g.NewNetwork(name); err != nil {
		t.Errorf("Failed to create network: %s", err)
		t.FailNow()
	}
	defer n.Close()

	if n.Name != name {
		t.FailNow()
	}

	if !hasNetwork(name) {
		t.FailNow()
	}

	if _, err := n.AddHost("h1"); err != nil {
		t.FailNow()
	}

	if _, err := netns.GetFromName(ns); err != nil {
		t.FailNow()
	}
}

func TestNetworkNSPrefix(t *testing.T) {
	var (
		err error
		n   *g.Network
	)

	prefix := "pfx-"
	name := g.GenerateNetworkName()
	ns := fmt.Sprintf("%s%s-h1", prefix, name)

	if n, err = g.NewNetwork(name, o.NSPrefix(prefix)); err != nil {
		t.Errorf("Failed to create network: %s", err)
		t.FailNow()
	}
	defer n.Close()

	if n.Name != name {
		t.FailNow()
	}

	if !hasNetwork(name) {
		t.FailNow()
	}

	if _, err := n.AddHost("h1"); err != nil {
		t.FailNow()
	}

	if _, err := netns.GetFromName(ns); err != nil {
		t.FailNow()
	}
}

func TestNetworkGeneratedName(t *testing.T) {
	var (
		err    error
		n1, n2 *g.Network
	)

	if n1, err = g.NewNetwork(""); err != nil {
		t.Errorf("Failed to create network: %s", err)
		t.FailNow()
	}
	defer n1.Close()

	if n2, err = g.NewNetwork(""); err != nil {
		t.FailNow()
	}
	defer n2.Close()
}

func TestNetworkExists(t *testing.T) {
	var (
		err    error
		n1, n2 *g.Network
	)

	name := g.GenerateNetworkName()

	if n1, err = g.NewNetwork(name); err != nil {
		t.Errorf("Failed to create network: %s", err)
		t.FailNow()
	}
	defer n1.Close()

	if n2, err = g.NewNetwork(name); err == nil {
		defer n2.Close()
		t.Errorf("Failed to create network: %s", err)
		t.FailNow()
	}
}
