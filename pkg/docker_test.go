package gont_test

import (
	"net/url"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

func TestDocker(t *testing.T) {
	// Test is broken
	t.Skip()

	var (
		err    error
		n      *g.Network
		sw     *g.Switch
		h1, h2 *g.Host
	)

	if n, err = g.NewNetwork(nname, opts...); err != nil {
		t.Errorf("Failed to create network: %s", err)
		t.FailNow()
	}
	defer n.Close()

	if sw, err = n.AddSwitch("sw"); err != nil {
		t.Errorf("Failed to create switch: %s", err)
		t.FailNow()
	}

	// h1 is a normal Gont node
	if h1, err = n.AddHost("h1",
		o.Interface("veth0", sw,
			o.AddressIPv4(10, 0, 0, 1, 24),
			o.AddressIP("fc::1/64")),
	); err != nil {
		t.Errorf("Failed to create host: %s", err)
		t.FailNow()
	}

	// h2 is a Docker container
	outp, _, err := n.HostNode.Run("docker", "run", "--detach", "nginx")
	if err != nil {
		t.Errorf("Failed to start Docker container")
		t.FailNow()
	}

	id := strings.TrimSpace(string(outp))

	log.WithField("id", id).Info("Started nginx Docker container")

	if h2, err = n.AddHost("h2",
		o.ExistingDockerContainer(id),
		o.Interface("veth0", sw,
			o.AddressIPv4(10, 0, 0, 2, 24),
			o.AddressIP("fc::2/64")),
	); err != nil {
		t.Errorf("Failed to create host: %s", err)
		t.FailNow()
	}

	h2.Run("hostname")

	u, err := url.Parse("http://h2/")
	if err != nil {
		t.Fail()
	}

	h1.Run("curl", u)
	h2.Ping(h1)
}
