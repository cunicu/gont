package gont_test

import (
	"testing"

	gont "github.com/stv0g/gont/pkg"
)

func TestRun(t *testing.T) {
	n := gont.NewNetwork("")

	_, _, err := n.Run("true")
	if err != nil {
		t.Fail()
	}

	_, _, err = n.Run("false")
	if err == nil {
		t.Fail()
	}

	ns, err := n.CreateNamespace("test")
	if err != nil {
		t.Fail()
	}

	_, _, err = n.RunNS(ns, "ip", "netns", "identify")
	if err != nil {
		t.Fail()
	}
}
