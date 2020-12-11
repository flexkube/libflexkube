FROM golang:1.15.6-alpine-3.12

# Install dependencies:
# - docker for spawning further Docker containers
# - curl for installing CI tools
# - build-base for make and gcc, required for running unit tests
# - bash and bash-completion for usable shell
# - vim for modifying files
RUN apk add -U docker curl build-base bash bash-completion vim
