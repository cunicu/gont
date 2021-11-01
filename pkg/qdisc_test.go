package gont_test

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-ping/ping"
	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

func testNetem(t *testing.T, ne o.Netem) (*ping.Statistics, error) {
	var (
		err    error
		n      *g.Network
		h1, h2 *g.Host
	)

	if n, err = g.NewNetwork(nname, opts...); err != nil {
		t.Errorf("Failed to create network: %s", err)
		t.FailNow()
	}
	defer n.Close()

	if h1, err = n.AddHost("h1"); err != nil {
		t.Errorf("Failed to create host: %s", err)
		t.FailNow()
	}

	if h2, err = n.AddHost("h2"); err != nil {
		t.Errorf("Failed to create host: %s", err)
		t.FailNow()
	}

	if err := n.AddLink(
		o.Interface("veth0", h1, ne,
			o.AddressIPv4(10, 0, 0, 1, 24)),
		o.Interface("veth0", h2,
			o.AddressIPv4(10, 0, 0, 2, 24)),
	); err != nil {
		t.Errorf("Failed to connect hosts: %s", err)
		t.FailNow()
	}

	return h1.PingWithOptions(h2, "ip", 1000, 2000*time.Millisecond, time.Millisecond, false)
}

// TestPingNetem performs and end-to-end ping test between two
// directly connected hosts via a link with a netem qdisc configured
//
// h1 <-> h2
func TestNetemLatency(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		// GitHubs Azure based CI environment is to unreliable
		// for this test to success consistently
		t.Skip()
	}

	latency := 50 * time.Millisecond

	ne := o.WithNetem(
		o.Latency(latency),
	)

	stats, err := testNetem(t, ne)
	if err != nil {
		t.Errorf("Failed to ping: %s", err)
	}

	t.Logf("AvgRtt: %s", stats.AvgRtt)

	diff := stats.AvgRtt - latency
	if diff < 0 {
		diff *= -1
	}

	if diff > 10*time.Millisecond {
		t.Fail()
	}
}

func TestNetemLoss(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		// GitHubs Azure based CI environment is to unreliable
		// for this test to success consistently
		t.Skip()
	}

	ne := o.WithNetem(
		o.Loss{Probability: 10.0},
	)

	stats, err := testNetem(t, ne)
	if err != nil && !strings.Contains(err.Error(), "lost") {
		t.Errorf("Failed to ping: %s", err)
	}

	t.Logf("Loss: %f", stats.PacketLoss)

	if math.Abs(stats.PacketLoss-float64(ne.Loss)) > 20 {
		t.Fail()
	}
}

func TestNetemDuplication(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		// GitHubs Azure based CI environment is to unreliable
		// for this test to success consistently
		t.Skip()
	}

	ne := o.WithNetem(
		o.Duplicate{Probability: 50.0},
	)

	stats, err := testNetem(t, ne)
	if err != nil {
		t.Errorf("Failed to ping: %s", err)
	}

	duplicatePercentage := 100.0 * float64(stats.PacketsRecvDuplicates) / float64(stats.PacketsSent)

	t.Logf("Duplicate packets: %d", stats.PacketsRecvDuplicates)
	t.Logf("Duplicate percentage: %.2f %%", duplicatePercentage)

	if math.Abs(duplicatePercentage-float64(ne.Duplicate)) > 10 {
		t.Fail()
	}
}
