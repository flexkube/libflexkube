# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test -covermode=atomic -buildmode=exe -v
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOBUILD=$(GOCMD) build -v -buildmode=exe

all: test lint build

build:
	$(GOBUILD)

test:
	$(GOTEST) ./...

test-race:
	$(GOTEST) -race ./...

test-integration:
	$(GOTEST) -tags=integration ./...

lint:
	which golangci-lint 2>&1 >/dev/null && golangci-lint run || echo "'golangci-lint' binary not found, skipping linting."
	which golint 2>&1 >/dev/null && golint -set_exit_status $$(go list ./...) || echo "'golint' binary not found, skipping linting."

update:
	$(GOGET) -u
	$(GOMOD) tidy
