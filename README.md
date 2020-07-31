<a href="https://www.cncf.io/certification/software-conformance/"><img alt="Certified Kubernetes logo" width="100px" align="right" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/kubernetes/certified-kubernetes/versionless/pantone/certified-kubernetes-pantone.png"></a>
<img alt="Flexkube logo" width="100px" align="right" src="https://github.com/flexkube/assets/raw/master/logo.jpg">

# libflexkube: Go library for deploying Kubernetes

[![Build Status](https://travis-ci.org/flexkube/libflexkube.svg?branch=master)](https://travis-ci.org/flexkube/libflexkube) [![Maintainability](https://api.codeclimate.com/v1/badges/5840c3fe0a9bc77aef08/maintainability)](https://codeclimate.com/github/flexkube/libflexkube/maintainability) [![codecov](https://codecov.io/gh/flexkube/libflexkube/branch/master/graph/badge.svg)](https://codecov.io/gh/flexkube/libflexkube) [![GoDoc](https://godoc.org/github.com/flexkube/libflexkube?status.svg)](https://godoc.org/github.com/flexkube/libflexkube) [![Go Report Card](https://goreportcard.com/badge/github.com/flexkube/libflexkube)](https://goreportcard.com/report/github.com/flexkube/libflexkube)

## Table of contents

- [Introduction](#introduction)
- [Documentation](#documentation)
- [Installation and usage](#installation-and-usage)
  - [CLI tool](#cli-tool)
  - [Terraform](#terraform)
  - [Next steps](#next-steps)
- [Characteristics](#characteristics)
- [Features](#features)
- [Known issues, missing features and limitations](#known-issues-missing-features-and-limitations)
- [Contributing](#contributing)
- [Status of the project](#status-of-the-project)

## Introduction

libflexkube is a core part of Flexkube project. It is a go library, which implements the logic for managing Kubernetes cluster components (e.g. [etcd](https://etcd.io/), [kubelet](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/), [certificates](https://kubernetes.io/docs/setup/best-practices/certificates/)) and provides reference implementation of [flexkube](https://flexkube.github.io/documentation/reference/cli/) CLI tool and [Terraform provider](https://flexkube.github.io/documentation/reference/terraform/).

Flexkube is a minimalistic and flexible Kubernetes distribution, providing tools to manage each main Kubernetes component independently, which can be used together to create Kubernetes clusters.

## Documentation

To see user documentation, visit [flexkube.github.io](https://flexkube.github.io). Please note, that the documentation is still in progress.

For development documentation, see [docs](docs).

## Installation and usage

For quick examples of installation and usage, see the content below. For full documentation, see [Getting started](https://flexkube.github.io/documentation/getting-started/).

### CLI tool

If you have `go` binary available in your system, you can start using Flexkube for creating Kubernetes certificates just by running the following commands:
```sh
echo 'pki:
  kubernetes: {}
  etcd: {}' > config.yaml
go run github.com/flexkube/libflexkube/cmd/flexkube pki
```

It will create `config.yaml` configuration file which will be consumed by `flexkube` CLI tool, which will then generate the certificates into `state.yaml` file.

### Terraform

If you want to perform the same action using Terraform, execute the following commands:

```sh
VERSION=v0.3.1 wget -qO- https://github.com/flexkube/libflexkube/releases/download/$VERSION/terraform-provider-flexkube_$VERSION_linux_amd64.tar.gz | tar zxvf - terraform-provider-flexkube_$VERSION_x4
echo 'resource "flexkube_pki" "pki" {
  etcd {}
  kubernetes {}
}' > main.tf
terraform init && terraform apply
```

After applying, Kubernetes certificates should be available as Terraform resource attributes of `flexkube_pki.pki` resource.

### Next steps

For more detailed instructions of how to use Flexkube, see the user [guides](https://flexkube.github.io/documentation/guides).

## Characteristics

Flexkube project focuses on simplicity and tries to only do minimal amount of steps in order to get Kubernetes cluster running, while keeping the configuration flexible and extensible. It is also a good material for learning how Kubernetes cluster is set up, as each part of the cluster is managed independently and code contains a lot of comments why specific flag/option is needed and what purpose does it serve.

Parts of this project could possibly be used in other Kubernetes distributions or be used as a configuration reference, as setting up Kubernetes components requires many flags and configuration files to be created.

Flexkube do not manage infrastructure for running Kubernetes clusters, it must be provided by the user.

## Features

Here is the short list of Flexkube project features:

- Minimal host requirements - Use SSH connection forwarding for talking directly to the container runtime on remote machines for managing static containers and configuration files.
- Independent management of etcd, kubelets, static control plane and self-hosted components.
- All self-hosted control plane components managed using Helm 3 (e.g CoreDNS).
- 1st class support for Terraform provider for automation.
- No Public DNS or any other public discovery service is required for getting cluster up and running.
- Others:
  - etcd, kubelet and static control plane running as containers.
  - Self-hosted control plane.
  - Supported container runtimes:
    - Docker
  - Configuration via YAML or via Terraform.
  - Deployment using CLI tools or via Terraform.
  - HAProxy for load-balancing and fail-over between Kubernetes API servers.

## Known issues, missing features and limitations

As the project is still in the early stages, here is the list of major existing issues or missing features, which are likely to be implemented in the future:

- gracefully replacing CA certificates (if private key does not change, it should work, but has not been tested)
- no checkpointer for pods/apiserver. If static kube-apiserver container is stopped and node reboots, single node cluster will not come back.

And features, which are not yet implemented:

- network policies for kube-system namespace
- caching port forwarding
- bastion host(s) support for SSH
- parallel deployments across hosts
- removal of configuration files, created data and containers
- automatic shutdown/start of bootstrap control plane

## Contributing

All contributions to this project are welcome. If it does not satisfy your needs, feel free to raise an issue about it or implement the support yourself and create a pull request with the patch, so we can all benefit from it.

If you just want to help the project grow and mature, there are many TODOs spread across the code, which should be addresses sooner or later.

## Status of the project

At the time of writing, this project is in active development state and it is not suitable for production use. Breaking changes might be introduced at any point in time, for both library interface and for existing deployments.
