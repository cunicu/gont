// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"flag"
	"os"
	"testing"

	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	co "github.com/stv0g/gont/pkg/options/capture"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//nolint:gochecknoglobals
var (
	globalNetworkOptions = []g.NetworkOption{}
	nname                = flag.String("name", "", "Network name")
	persist              = flag.Bool("persist", false, "Do not teardown networks after test")
	capture              = flag.String("capture", "", "Capture network traffic to PCAPng file")
)

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

func TestMain(m *testing.M) {
	logger := setupLogging()
	defer logger.Sync() //nolint:errcheck

	// Handle global flags
	if *persist {
		globalNetworkOptions = append(globalNetworkOptions,
			o.Persistent(*persist),
		)
	}

	if *capture != "" {
		globalNetworkOptions = append(globalNetworkOptions,
			g.NewCapture(
				co.Filename(*capture)),
		)
	}

	os.Exit(m.Run())
}
