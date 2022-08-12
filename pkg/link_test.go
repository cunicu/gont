package gont_test

import (
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

func TestLink(t *testing.T) {
	var (
		err    error
		n      *g.Network
		h1, h2 *g.Host
	)

	if n, err = g.NewNetwork(*nname, globalNetworkOptions...); err != nil {
		t.FailNow()
	}
	defer n.Close()

	if h1, err = n.AddHost("h1"); err != nil {
		t.FailNow()
	}

	if h2, err = n.AddHost("h2"); err != nil {
		t.FailNow()
	}

	if err = n.AddLink(
		o.Interface("veth0", h1),
		o.Interface("veth0", h2),
	); err != nil {
		t.Errorf("Failed to link nodes: %s", err)
	}
}
