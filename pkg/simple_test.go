package gont_test

import (
	"net"
	"testing"

	gont "github.com/stv0g/gont/pkg"
)

func mask() net.IPMask {
	return net.IPv4Mask(255, 255, 255, 0)
}

// TestPing performs and end-to-end ping test
// between two hosts on a switched topology
//
//  h1 <-> sw <-> h2
func TestPing(t *testing.T) {
	n := gont.NewNetwork("")
	defer n.Close()

	sw, err := n.AddSwitch("sw")
	if err != nil {
		t.Fail()
	}

	h1, err := n.AddHost("h1", nil, &gont.Interface{"eth0", net.IPv4(10, 0, 0, 1), mask(), sw})
	if err != nil {
		t.Fail()
	}

	h2, err := n.AddHost("h2", nil, &gont.Interface{"eth0", net.IPv4(10, 0, 0, 2), mask(), sw})
	if err != nil {
		t.Fail()
	}

	if err := gont.TestConnectivity(h1, h2); err != nil {
		t.Fail()
	}
}

// TestPingMultiHop performs and end-to-end ping test
// through a routed multi-hop topology
//
//  h1 <-> sw1 <-> r1 <-> sw2 <-> h2
func TestPingMultiHop(t *testing.T) {
	n := gont.NewNetwork("")
	defer n.Close()

	sw1, err := n.AddSwitch("sw1")
	if err != nil {
		t.Fail()
	}

	sw2, err := n.AddSwitch("sw2")
	if err != nil {
		t.Fail()
	}

	h1, err := n.AddHost("h1", net.IPv4(10, 0, 1, 1), &gont.Interface{"eth0", net.IPv4(10, 0, 1, 2), mask(), sw1})
	if err != nil {
		t.Fail()
	}

	h2, err := n.AddHost("h2", net.IPv4(10, 0, 2, 1), &gont.Interface{"eth0", net.IPv4(10, 0, 2, 2), mask(), sw2})
	if err != nil {
		t.Fail()
	}

	_, err = n.AddRouter("r1", nil,
		&gont.Interface{"eth0", net.IPv4(10, 0, 1, 1), mask(), sw1},
		&gont.Interface{"eth1", net.IPv4(10, 0, 2, 1), mask(), sw2})
	if err != nil {
		t.Fail()
	}

	if err := gont.TestConnectivity(h1, h2); err != nil {
		t.Fail()
	}
}
