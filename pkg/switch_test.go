package gont_test

import (
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

// TestPing performs and end-to-end ping test
// between two hosts on a switched topology
//
//  h1 <-> sw1 <-> sw2 <-> h2
func TestPingCascadedSwitches(t *testing.T) {
	var (
		err      error
		n        *g.Network
		sw1, sw2 *g.Switch
		h1, h2   *g.Host
	)

	if n, err = g.NewNetwork(nname, opts...); err != nil {
		t.Errorf("Failed to create network: %s", err)
		t.FailNow()
	}
	defer n.Close()

	if sw1, err = n.AddSwitch("sw1"); err != nil {
		t.Errorf("Failed to add switch: %s", err)
		t.FailNow()
	}

	if sw2, err = n.AddSwitch("sw2"); err != nil {
		t.Errorf("Failed to add switch: %s", err)
		t.FailNow()
	}

	if h1, err = n.AddHost("h1",
		o.Interface("veth0", sw1,
			o.AddressIPv4(10, 0, 0, 1, 24)),
	); err != nil {
		t.Errorf("Failed to add host: %s", err)
		t.FailNow()
	}

	if h2, err = n.AddHost("h2",
		o.Interface("veth0", sw2,
			o.AddressIPv4(10, 0, 0, 2, 24)),
	); err != nil {
		t.Errorf("Failed to add host: %s", err)
		t.FailNow()
	}

	if err = n.AddLink(
		o.Interface("br-sw2", sw1),
		o.Interface("br-sw1", sw2),
	); err != nil {
		t.Errorf("Failed to add link: %s", err)
		t.FailNow()
	}

	if err := g.TestConnectivity(h1, h2); err != nil {
		t.Errorf("Failed to check connectivity: %s", err)
	}
}
