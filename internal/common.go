package internal

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func SetupLogging() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(logger)
	zap.LevelFlag("log-level", zap.InfoLevel, "Log level")

	return logger
}

func SetupRand() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func SetupSignals() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	return ch
}
