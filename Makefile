BINARY=bin/db-backup
PKG := github.com/fidrasofyan/db-backup/internal
VERSION := $(shell git describe --tags --always --dirty)
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X '$(PKG)/tasks.AppVersion=$(VERSION)' \
	-X '$(PKG)/tasks.AppCommit=$(COMMIT)' \
	-X '$(PKG)/tasks.AppBuildTime=$(DATE)'

build:
	go vet ./...
	staticcheck ./...
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd
	