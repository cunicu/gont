package gont_test

import (
	"flag"
	"io"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
)

var opts = g.Options{}
var nname string

func TestMain(m *testing.M) {
	g.SetupRand()
	g.SetupLogging()

	var persist bool
	flag.BoolVar(&persist, "persist", false, "Do not teardown networks after test")
	flag.StringVar(&nname, "name", "", "Network name")
	flag.Parse()

	if persist {
		opts = append(opts, o.Persistent(persist))
	}

	if !testing.Verbose() {
		log.SetOutput(io.Discard)
	}

	os.Exit(m.Run())
}
