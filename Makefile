# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

GIT_TAG = $(shell git describe --tags HEAD)

GOFLAGS = -ldflags "-X main.tag=$(GIT_TAG)"

all: gontc

gontc:
	go build $(GOFLAGS) -o $@ ./cmd/gontc

.PHONY: all gontc
