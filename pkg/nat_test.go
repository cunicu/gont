package gont_test

import (
	"net"
	"testing"

	gont "github.com/stv0g/gont/pkg"
)

// TestPingNAT performs and end-to-end ping test
// through a NAT topology
//
//  h1 <-> sw1 <-> nat1 <-> sw2 <-> h2
func TestPingNAT(t *testing.T) {
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

	_, err = n.AddNAT("n1", nil,
		&gont.Interface{"eth0", net.IPv4(10, 0, 1, 1), mask(), sw1},
		&gont.Interface{"eth1", net.IPv4(10, 0, 2, 1), mask(), sw2})
	if err != nil {
		t.Fail()
	}

	err = h1.Ping(h2, "-c", "1")
	if err != nil {
		t.Fail()
	}

	err = h1.Traceroute(h2)
	if err != nil {
		t.Fail()
	}
}

// TestPingDoubleNAT performs and end-to-end ping test
// through a double / carrier-grade NAT topology
//
//  h1 <-> sw1 <-> nat1 <-> sw2 <-> nat2 <-> sw3 <-> h2
func TestPingDoubleNAT(t *testing.T) {
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

	sw3, err := n.AddSwitch("sw3")
	if err != nil {
		t.Fail()
	}

	h1, err := n.AddHost("h1", net.IPv4(10, 0, 1, 1), &gont.Interface{"eth0", net.IPv4(10, 0, 1, 2), mask(), sw1})
	if err != nil {
		t.Fail()
	}

	h2, err := n.AddHost("h2", net.IPv4(10, 0, 2, 1), &gont.Interface{"eth0", net.IPv4(10, 0, 2, 2), mask(), sw3})
	if err != nil {
		t.Fail()
	}

	_, err = n.AddNAT("n1", net.IPv4(10, 0, 3, 1),
		&gont.Interface{"eth0", net.IPv4(10, 0, 3, 2), mask(), sw2},
		&gont.Interface{"eth1", net.IPv4(10, 0, 1, 1), mask(), sw1})
	if err != nil {
		t.Fail()
	}

	_, err = n.AddNAT("n2", nil,
		&gont.Interface{"eth0", net.IPv4(10, 0, 2, 1), mask(), sw3},
		&gont.Interface{"eth1", net.IPv4(10, 0, 3, 1), mask(), sw2})
	if err != nil {
		t.Fail()
	}

	err = h1.Ping(h2, "-c", "1")
	if err != nil {
		t.Fail()
	}

	err = h1.Traceroute(h2)
	if err != nil {
		t.Fail()
	}
}

// TestPingHostNAT performs and end-to-end ping test
// between a virtual host and the outside host network
//
//  h1 <-> sw <-> nat1 <-> external
// func TestPingHostNAT(t *testing.T) {
// 	n := gont.NewNetwork("")
// 	defer n.Close()

// 	sw, err := n.AddSwitch("sw")
// 	if err != nil {
// 		t.Fail()
// 	}

// 	h, err := n.AddHost("h", nil, &gont.Interface{"eth0", net.IPv4(10, 0, 0, 1), mask(), sw})
// 	if err != nil {
// 		t.Fail()
// 	}

// 	_, err = h.Run("ping", "-c", "1", "1.1.1.1")
// 	if err != nil {
// 		t.Fail()
// 	}
// }
