// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"

	"cunicu.li/gont/v2/pkg/trace"
)

func main() {
	if err := trace.Start(0); err != nil {
		log.Printf("Failed to start tracer: %s", err)
		return
	}

	myData := map[string]any{
		"Hello": "World",
	}

	if err := trace.PrintfWithData(myData, "This is my first trace message: %s", "Hurra"); err != nil {
		log.Printf("Failed to write trace: %s", err)
	}

	if err := trace.Stop(); err != nil {
		log.Printf("Failed to stop trace: %s", err)
	}
}
