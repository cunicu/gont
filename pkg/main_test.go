package gont_test

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:  true,
		DisableQuote: true,
	})

	os.Exit(m.Run())
}
