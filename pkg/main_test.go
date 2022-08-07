package gont_test

import (
	crand "crypto/rand"
	"encoding/binary"
	"flag"
	"math/rand"
	"os"
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalOpts = g.Options{}
var nname = flag.String("name", "", "Network name")
var persist = flag.Bool("persist", false, "Do not teardown networks after test")

func setupLogging() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()

	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.99")

	logger, err := cfg.Build()
	if err != nil {
		panic("failed to setup logging")
	}

	zap.ReplaceGlobals(logger)
	zap.LevelFlag("log-level", zap.DebugLevel, "Log level")

	return logger
}

func setupRand() error {
	var seed int64

	if err := binary.Read(crand.Reader, binary.LittleEndian, &seed); err != nil {
		return err
	}

	rand.Seed(seed)

	return nil
}

func TestMain(m *testing.M) {
	setupRand()
	logger := setupLogging()
	defer logger.Sync()

	// Handle global flags
	if *persist {
		globalOpts = append(globalOpts, o.Persistent(*persist))
	}

	os.Exit(m.Run())
}
