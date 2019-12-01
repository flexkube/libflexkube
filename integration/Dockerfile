FROM golang:1.13-alpine

# Install dependencies:
# - docker for spawning further Docker containers
# - curl for installing CI tools
# - build-base for make and gcc, required for running unit tests
RUN apk add -U docker curl build-base
