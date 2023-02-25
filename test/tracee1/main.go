// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"

	"github.com/stv0g/gont/pkg/trace"
)

func main() {
	if err := trace.Start(0); err != nil {
		log.Fatalf("Failed to start tracer: %s", err)
	}
	defer trace.Stop() //nolint:errcheck

	myData := map[string]any{
		"Hello": "World",
	}

	if err := trace.PrintfWithData(myData, "This is my first trace message: %s", "Hurra"); err != nil {
		log.Fatalf("Failed to write trace: %s", err)
	}
}
