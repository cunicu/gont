package topo_test

import (
	"testing"

	g "github.com/stv0g/gont/pkg"
	"github.com/stv0g/gont/pkg/topo"
)

func TestSingleSwitch(t *testing.T) {
	t.Skip()

	n, err := g.NewNetwork("")
	if err != nil {
		t.Errorf("Failed to create network: %s", err)
		t.FailNow()
	}

	_, hs, err := topo.SingleSwitch(n, 16)
	if err != nil {
		t.Fail()
	}

	if err := g.TestConnectivity(hs...); err != nil {
		t.Fail()
	}
}
