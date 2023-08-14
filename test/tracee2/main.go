// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"cunicu.li/gont/v2/pkg/trace"
	"go.uber.org/zap"
)

func main() {
	if err := trace.Start(0); err != nil {
		panic(err)
	}
	defer trace.Stop() //nolint:errcheck

	traced()
	log()
	ping()
}

func traced() {
	ts := time.Now()

	fmt.Printf("My time is: %s\n", ts)

	data := map[string]any{
		"i": 1337,
	}

	if err := trace.PrintfWithData(data, "My time is: %s\n", ts); err != nil {
		fmt.Println(err)
	}
}

func log() {
	logger := zap.L().Named("log")

	logger.Named("my_first_logger").Debug("Debug")
	logger.Named("my_second_logger").Warn("Warning")
	logger.Named("my_third_logger").Info("Info")
	logger.Named("my_fourth_logger").Error("Error")

	logger.Info("This is a test",
		zap.String("hallo", "welt"),
		zap.Any("any", map[string]any{
			"a": 1,
			"b": false,
			"c": nil,
			"d": map[string]any{
				"1": 1 * time.Hour,
				"2": []int{1, 2, 3, 4},
			},
		}),
		zap.Time("in_one_year", time.Now().Add(24*365*time.Hour)),
	)
}

func ping() {
	logger := zap.L().Named("ping")

	for i := 0; i < 5; i++ {
		start := time.Now()
		cmd := exec.Command("ping", "-c", "1", "127.0.0.1")
		cmd.Stdout = os.Stdout
		err := cmd.Run()

		elapsed := time.Since(start)

		logger.Info("Pinged", zap.Duration("rtt", elapsed), zap.Error(err))

		time.Sleep(100 * time.Millisecond)
	}
}
