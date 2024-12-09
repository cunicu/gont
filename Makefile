# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

GIT_TAG = $(shell git describe --tags HEAD)

export GOFLAGS = -buildvcs=false -ldflags=-X=main.tag=$(GIT_TAG)

all: gontc

tests:
	sudo -E go test ./pkg -v $(TEST_OPTS)
	sudo -E go test ./internal -v $(TEST_OPTS)

gontc:
	go build -o $@ ./cmd/gontc

.PHONY: all gontc
