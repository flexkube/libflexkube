# Build parameters
CGO_ENABLED=0
LD_FLAGS="-extldflags '-static'"

# Go parameters
GOCMD=env GO111MODULE=on go
GOTEST=$(GOCMD) test -covermode=atomic -buildmode=exe -v
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOBUILD=CGO_ENABLED=$(CGO_ENABLED) $(GOCMD) build -v -buildmode=exe -ldflags $(LD_FLAGS)

CC_TEST_REPORTER_ID=6e107e510c5479f40b0ce9166a254f3f1ee0bc547b3e48281bada1a5a32bb56d
GOLANGCI_LINT_VERSION=v1.21.0
BIN_PATH=$$HOME/bin

GO_PACKAGES=./...

INTEGRATION_IMAGE=flexkube/libflexkube-integration

INTEGRATION_CMD=docker run -it --rm -v /run:/run -v /home/core/libflexkube:/usr/src/libflexkube -v /home/core/go:/go -v /home/core/.password:/home/core/.password -v /home/core/.ssh:/home/core/.ssh -v /home/core/.cache:/root/.cache -w /usr/src/libflexkube --net host $(INTEGRATION_IMAGE)

DISABLED_LINTERS=golint,godox,lll,funlen,dupl,gocyclo,gocognit,gosec

.PHONY: all
all: build test lint

.PHONY: all-cover
all-cover: build test-cover lint

.PHONY: build
build:
	$(GOBUILD) ./cmd/...

.PHONY: build-bin
build-bin:
	mkdir -p ./bin
	cd bin && for i in $$(ls ../cmd); do $(GOBUILD) ../cmd/$$i; done

.PHONY: build-docker
build-docker:
	docker build .

.PHONY: clean
clean:
	rm -r ./bin c.out coverage.txt 2>/dev/null || true

.PHONY: test
test:
	$(GOTEST) $(GO_PACKAGES)

.PHONY: download
download:
	$(GOMOD) download

.PHONY: test-race
test-race:
	$(GOTEST) -race $(GO_PACKAGES)

.PHONY: test-integration
test-integration:
	$(GOTEST) -tags=integration $(GO_PACKAGES)

.PHONY: test-cover
test-cover:
	$(GOTEST) -coverprofile=$(PROFILEFILE) $(GO_PACKAGES)

.PHONY: lint
lint:
	golangci-lint run --enable-all --disable=$(DISABLED_LINTERS) --max-same-issues=0 --max-issues-per-linter=0 --build-tags integration $(GO_PACKAGES)
	# Since golint is very opinionated about certain things, for example exported functions returning
	# unexported structs, which we use here a lot, let's filter them out and set status ourselves.
	#
	# TODO Maybe cache golint result somewhere, do we don't have to run it twice?
	golint $$(go list $(GO_PACKAGES)) | grep -v -E 'returns unexported type.*, which can be annoying to use' || true
	test $$(golint $$(go list $(GO_PACKAGES)) | grep -v -E "returns unexported type.*, which can be annoying to use" | wc -l) -eq 0

.PHONY: update
update:
	$(GOGET) -u $(GO_PACKAGES)
	$(GOMOD) tidy

.PHONY: codespell
codespell:
	codespell -S .git,state.yaml,go.sum,terraform.tfstate

.PHONY: codespell-pr
codespell-pr:
	git diff master..HEAD | grep -v ^- | codespell -
	git log master..HEAD | codespell -

.PHONY: format
format:
	goimports -l -w $$(find . -name '*.go' | grep -v '^./vendor')

.PHONY: codecov
codecov: PROFILEFILE=coverage.txt
codecov: test-cover
codecov:
	bash <(curl -s https://codecov.io/bash)

.PHONY: codeclimate-prepare
codeclimate-prepare:
	cc-test-reporter before-build
\
.PHONY: codeclimate
codeclimate: PROFILEFILE=c.out
codeclimate: codeclimate-prepare test-cover
codeclimate:
	env CC_TEST_REPORTER_ID=$(CC_TEST_REPORTER_ID) cc-test-reporter after-build --exit-code $(EXIT_CODE)

.PHONY: install-golangci-lint
install-golangci-lint:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(BIN_PATH) $(GOLANGCI_LINT_VERSION)

.PHONY: install-golint
install-golint:
	$(GOGET) -u golang.org/x/lint/golint

.PHONY: install-cc-test-reporter
install-cc-test-reporter:
	curl -L https://codeclimate.com/downloads/test-reporter/test-reporter-latest-linux-amd64 > $(BIN_PATH)/cc-test-reporter
	chmod +x $(BIN_PATH)/cc-test-reporter

.PHONY: install-ci
install-ci: install-golangci-lint install-golint install-cc-test-reporter

.PHONY: vagrant-up
vagrant-up:
	vagrant up

.PHONY: vagrant-rsync
vagrant-rsync:
	vagrant rsync

.PHONY: vagrant-destroy
vagrant-destroy:
	vagrant destroy --force

.PHONY: vagrant-integration-build
vagrant-integration-build:
	vagrant ssh -c "cd libflexkube && docker build -t $(INTEGRATION_IMAGE) integration"

.PHONY: vagrant-integration-run
vagrant-integration-run:
	vagrant ssh -c "$(INTEGRATION_CMD) make test-integration GO_PACKAGES=$(GO_PACKAGES)"

.PHONY: vagrant-integration-shell
vagrant-integration-shell:
	vagrant ssh -c "$(INTEGRATION_CMD) bash"

.PHONY: vagrant-integration
vagrant-integration: vagrant-up vagrant-rsync vagrant-integration-build vagrant-integration-run
