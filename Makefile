# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test -covermode=atomic -buildmode=exe -v
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOBUILD=$(GOCMD) build -v -buildmode=exe

all: test lint build

build:
	$(GOBUILD) ./cmd/...

test:
	$(GOTEST) ./...

test-race:
	$(GOTEST) -race ./...

test-integration:
	$(GOTEST) -tags=integration ./...

lint:
	golangci-lint run --enable-all --disable=golint,godox,lll,funlen,dupl,gocyclo,gocognit,gosec
	# Since golint is very opinionated about certain things, for example exported functions returning
	# unexported structs, which we use here a lot, let's filter them out and set status ourselves.
	#
	# TODO Maybe cache golint result somewhere, do we don't have to run it twice?
	golint $$(go list ./...) | grep -v -E 'returns unexported type.*, which can be annoying to use' || true
	test $$(golint $$(go list ./...) | grep -v -E "returns unexported type.*, which can be annoying to use" | wc -l) -eq 0

update:
	$(GOGET) -u
	$(GOMOD) tidy

codespell:
	codespell  -S .git,state.yaml,go.sum,terraform.tfstate

codespell-pr:
	git diff master..HEAD | grep -v ^- | codespell -
	git log master..HEAD | codespell -

format:
	gofmt -s -l -w $$(find . -name '*.go' | grep -v '^./vendor')
	goimports -l -w $$(find . -name '*.go' | grep -v '^./vendor')
