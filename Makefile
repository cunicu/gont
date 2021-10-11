GIT_COMMIT = $(shell git rev-list -1 HEAD)
GIT_TAG = $(shell git describe --tags HEAD)

GOFLAGS = -ldflags "-X main.GitCommit=$(GIT_COMMIT) -X main.GitTag=$(GIT_TAG)"

all: gontc

gontc:
	go build $(GOFLAGS) -o $@ cmd/gontc/*.go

.PHONY: all gontc
