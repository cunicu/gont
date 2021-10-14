package gont_test

import (
	"testing"
	"time"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

func prepareQdisc(t *testing.T) (*g.Network, *g.Host, *g.Host) {
	var (
		err    error
		n      *g.Network
		h1, h2 *g.Host
	)

	if n, err = g.NewNetwork(nname, opts...); err != nil {
		t.Errorf("Failed to create network: %s", err)
		t.FailNow()
	}

	if h1, err = n.AddHost("h1"); err != nil {
		t.Errorf("Failed to create host: %s", err)
		t.FailNow()
	}

	if h2, err = n.AddHost("h2"); err != nil {
		t.Errorf("Failed to create host: %s", err)
		t.FailNow()
	}

	return n, h1, h2
}

// TestPingNetem performs and end-to-end ping test between two
// directly connected hosts via a link with a netem qdisc configured
//
// h1 <-> h2
func TestPingNetem(t *testing.T) {
	n, h1, h2 := prepareQdisc(t)
	defer n.Close()

	netem := o.WithNetem(
		o.Latency(10 * time.Millisecond),
	)

	tbf := o.WithTbf(
		o.Rate(200000),
	)

	if err := n.AddLink(
		o.Interface("veth0", h1, netem, tbf,
			o.AddressIPv4(10, 0, 0, 1, 24)),
		o.Interface("veth0", h2, netem, tbf,
			o.AddressIPv4(10, 0, 0, 2, 24)),
	); err != nil {
		t.Errorf("Failed to connect hosts: %s", err)
		t.FailNow()
	}

	h1.Run("ping", "-c", "1", "h2")
	_, _, e, err := h1.Start("iperf", "-s")
	if err != nil {
		t.Error(err)
	}

	if _, _, err = h2.Run("iperf", "-c", "h1"); err != nil {
		t.Error(err)
	}

	e.Process.Kill()
	e.Wait()
}
