// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	tco "cunicu.li/gont/v2/pkg/options/tc"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/stretchr/testify/require"
)

func testNetem(t *testing.T, ne o.Netem) (*probing.Statistics, error) {
	n, err := g.NewNetwork(*nname)
	require.NoError(t, err, "Failed to create network")
	defer n.MustClose()

	h1, err := n.AddHost("h1")
	require.NoError(t, err, "Failed to create host")

	h2, err := n.AddHost("h2")
	require.NoError(t, err, "Failed to create host")

	err = n.AddLink(
		g.NewInterface("veth0", h1, ne,
			o.AddressIP("10.0.0.1/24")),
		g.NewInterface("veth0", h2,
			o.AddressIP("10.0.0.2/24")))
	require.NoError(t, err, "Failed to connect hosts")

	return h1.PingWithOptions(h2, "ip", 1000, 2000*time.Millisecond, time.Millisecond, false)
}

// TestPingNetem performs and end-to-end ping test between two
// directly connected hosts via a link with a netem qdisc configured
//
// h1 <-> h2
func TestNetemLatency(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		t.Skip("GitHubs Azure based CI environment is too unreliable for this test to succeed consistently")
	}

	latency := 50 * time.Millisecond

	ne := o.WithNetem(
		tco.Latency(latency),
	)

	stats, err := testNetem(t, ne)
	require.NoError(t, err, "Failed to ping")

	t.Logf("AvgRtt: %s", stats.AvgRtt)

	diff := stats.AvgRtt - latency
	if diff < 0 {
		diff *= -1
	}

	require.Less(t, diff, 10*time.Millisecond, "Latency deviation too large")
}

func TestNetemLoss(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		t.Skip("GitHubs Azure based CI environment is too unreliable for this test to success consistently")
	}

	ne := o.WithNetem(
		tco.Loss{Probability: 10.0},
	)

	stats, err := testNetem(t, ne)
	require.False(t, err != nil && !strings.Contains(err.Error(), "lost"), "Failed to ping")

	t.Logf("Loss: %f", stats.PacketLoss)

	require.Less(t, math.Abs(stats.PacketLoss-float64(ne.Loss)), 25.0)
}

func TestNetemDuplication(t *testing.T) {
	if _, ok := os.LookupEnv("GITHUB_WORKFLOW"); ok {
		t.Skip("GitHubs Azure based CI environment is to unreliable for this test to success consistently")
	}

	ne := o.WithNetem(
		tco.Duplicate{Probability: 50.0},
	)

	stats, err := testNetem(t, ne)
	require.NoError(t, err, "Failed to ping")

	duplicatePercentage := 100.0 * float64(stats.PacketsRecvDuplicates) / float64(stats.PacketsSent)

	t.Logf("Duplicate packets: %d", stats.PacketsRecvDuplicates)
	t.Logf("Duplicate percentage: %.2f %%", duplicatePercentage)

	require.Less(t, math.Abs(duplicatePercentage-float64(ne.Duplicate)), 10.0)
}
