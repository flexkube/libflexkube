# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

all: test lint

test:
	$(GOTEST) -v ./...

lint:
	which golangci-lint 2>&1 >/dev/null && golangci-lint run || echo "'golangci-lint' binary not found, skipping linting."

update:
	$(GOGET) -u
	$(GOMOD) tidy
