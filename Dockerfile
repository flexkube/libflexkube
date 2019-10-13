FROM golang:1.12-alpine

# Enable go modules
ENV GO111MODULE=on

# Install dependencies
RUN apk add curl git build-base

# Install linter
RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $HOME/bin v1.17.1

# Copy go mod files first and install dependencies to cache this layer
ADD ./go.mod /usr/src/flexkube/
WORKDIR /usr/src/flexkube
RUN go get

# Add source code
ADD . /usr/src/flexkube

# Test and lint
RUN go test -v ./... && \
    $HOME/bin/golangci-lint run
