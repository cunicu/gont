// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont_test

import (
	"flag"
	"os"
	"testing"

	g "cunicu.li/gont/v2/pkg"
	o "cunicu.li/gont/v2/pkg/options"
	co "cunicu.li/gont/v2/pkg/options/capture"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//nolint:gochecknoglobals
var (
	nname   = flag.String("name", "", "Network name")
	persist = flag.Bool("persist", false, "Do not teardown networks after test")
	capture = flag.String("capture", "", "Capture network traffic to PCAPng file")
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
	flag.Parse()

	logger := setupLogging()

	// Handle global flags
	if *persist {
		g.GlobalOptions = append(g.GlobalOptions,
			o.Persistent(*persist),
		)
	}

	if *capture != "" {
		g.GlobalOptions = append(g.GlobalOptions,
			g.NewCapture(
				co.Filename(*capture)),
		)
	}

	rc := m.Run()

	logger.Sync() //nolint:errcheck

	os.Exit(rc)
}
