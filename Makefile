BINARY ?= dmp
PACKAGE := ./cmd/dmp
CGO_ENABLED ?= 0

VERSION ?= $(shell git describe --tags --always 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo "dev")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date +"%Y-%m-%d %H:%M:%S")

LDFLAGS := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.GitBranch=$(GIT_BRANCH)' \
	-X 'main.GitCommit=$(GIT_COMMIT)' \
	-X 'main.BuildTime=$(BUILD_TIME)'

.PHONY: test build version clean

test:
	go test ./...

build:
	CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o $(BINARY) $(PACKAGE)

version: build
	./$(BINARY) version

clean:
	rm -f $(BINARY)
