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
	golangci-lint run
	golint -set_exit_status $$(go list ./...)

update:
	$(GOGET) -u
	$(GOMOD) tidy
