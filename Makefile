# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test -v -covermode=atomic -buildmode=exe ./...
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOBUILD=$(GOCMD) build -v -buildmode=exe

all: test lint build

build:
	$(GOBUILD)

test:
	$(GOTEST)

test-race:
	$(GOTEST) -race

lint:
	which golangci-lint 2>&1 >/dev/null && golangci-lint run || echo "'golangci-lint' binary not found, skipping linting."

update:
	$(GOGET) -u
	$(GOMOD) tidy
