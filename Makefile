BINARY=bin/s3-backup
PKG := github.com/fidrasofyan/s3-backup/internal
VERSION := $(shell git describe --tags --always --dirty)
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X '$(PKG)/tasks.AppVersion=$(VERSION)' \
	-X '$(PKG)/tasks.AppCommit=$(COMMIT)' \
	-X '$(PKG)/tasks.AppBuildTime=$(DATE)'

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/s3-backup
