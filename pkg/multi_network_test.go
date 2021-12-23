package gont_test

import (
	"fmt"
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

func prepareNetwork(t *testing.T, i int) *g.Network {
	var (
		err error
		n   *g.Network
		sw  *g.Switch
	)

	pfx := fmt.Sprintf("net%d-", i)

	address := func(j int) o.Address {
		return o.AddressIP(fmt.Sprintf("fc::%d:%d/112", i, j))
	}

	if n, err = g.NewNetwork("", opts...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	if sw, err = n.AddSwitch(pfx + "sw"); err != nil {
		t.Fatalf("Failed to create switch: %s", err)
	}

	if _, err = n.AddHost(pfx+"h1",
		o.Interface("veth0", sw,
			address(1)),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	if _, err = n.AddHost(pfx+"h2",
		o.Interface("veth0", sw,
			address(2)),
	); err != nil {
		t.Fatalf("Failed to create host: %s", err)
	}

	return n
}

func TestMultipleNetworks(t *testing.T) {
	n1 := prepareNetwork(t, 1)
	defer n1.Close()

	n2 := prepareNetwork(t, 2)
	defer n2.Close()

	// Connectivity within the network must succeed
	if err := g.TestConnectivity(n1.Hosts()...); err != nil {
		t.Errorf("Connectivity tests between hosts on same network must succeed")
	}

	// Connectivity within the network must succeed
	if err := g.TestConnectivity(n2.Hosts()...); err != nil {
		t.Errorf("Connectivity tests between hosts on same network must succeed")
	}

	// Connectivity between the networks must fail
	all := append(n1.Hosts(), n2.Hosts()...)
	if err := g.TestConnectivity(all...); err == nil {
		t.Errorf("Connectivity tests between hosts on different networks must fail")
	}
}

func TestNetworkNameCollision(t *testing.T) {
	var (
		err    error
		n1, n2 *g.Network
	)

	if n1, err = g.NewNetwork("", opts...); err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}
	defer n1.Close()

	// Creating another network with the same name must fail
	if n2, err = g.NewNetwork(n1.Name, opts...); err == nil {
		defer n2.Close()
		t.Fatalf("Cannot create multiple networks with same name")
	}
}
