FROM golang:1.18-alpine3.15

# Enable go modules
ENV GO111MODULE=on

# Install dependencies
RUN apk add curl git build-base

# Copy Makefile first to install CI binaries etc.
ADD ./Makefile /usr/src/libflexkube/

WORKDIR /usr/src/libflexkube

RUN make install-ci BIN_PATH=/usr/local/bin

# Copy go mod files first and install dependencies to cache this layer
ADD ./go.mod ./go.sum /usr/src/libflexkube/

RUN make download

# Add source code
ADD . /usr/src/libflexkube

# Build, test and lint
RUN make all build-bin
