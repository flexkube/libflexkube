<a href="https://www.cncf.io/certification/software-conformance/"><img alt="Certified Kubernetes logo" width="100px" align="right" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/kubernetes/certified-kubernetes/versionless/pantone/certified-kubernetes-pantone.png"></a>
<img alt="Flexkube logo" width="100px" align="right" src="https://github.com/flexkube/assets/raw/master/logo.jpg">

# libflexkube: Go library for deploying Kubernetes

[![Build Status](https://travis-ci.org/flexkube/libflexkube.svg?branch=master)](https://travis-ci.org/flexkube/libflexkube) [![Maintainability](https://api.codeclimate.com/v1/badges/5840c3fe0a9bc77aef08/maintainability)](https://codeclimate.com/github/flexkube/libflexkube/maintainability) [![codecov](https://codecov.io/gh/flexkube/libflexkube/branch/master/graph/badge.svg)](https://codecov.io/gh/flexkube/libflexkube) [![GoDoc](https://godoc.org/github.com/flexkube/libflexkube?status.svg)](https://godoc.org/github.com/flexkube/libflexkube) [![Go Report Card](https://goreportcard.com/badge/github.com/flexkube/libflexkube)](https://goreportcard.com/report/github.com/flexkube/libflexkube)

## Table of contents

- [Introduction](#introduction)
- [Features](#features)
- [Requirements](#requirements)
- [User tools](#user-tools)
  * [CLI binaries](#cli-binaries)
  * [Terraform provider](#terraform-provider)
- [Example usage](#example-usage)
  * [Deploying over SSH](#deploying-over-ssh)
- [Supported container runtimes](#supported-container-runtimes)
- [Supported transport protocols](#supported-transport-protocols)
- [Managing certificates](#managing-certificates)
- [Getting started](#getting-started)
- [Current known issues and limitations](#current-known-issues-and-limitations)
- [Testing](#testing)
- [Helm charts](#helm-charts)
- [Contributing](#contributing)
- [Status of the project](#status-of-the-project)

## Introduction

libflexkube is a go library, which implements the logic required for deploying self-hosted Kubernetes cluster.

The purpose of this project is to provide generic boilerplate code required for running Kubernetes cluster. Each part of this library is an independent piece,
which can be mixed with other projects and other Kubernetes distributions.

For example, if you don't like how your distribution runs etcd, you can use only etcd management part of this library.

Here is high-level overview of what packages are provided:
- [pkg/container](pkg/container) - allows to manage containers using different container runtimes across multiple hosts
- [pkg/host](pkg/container) - allows to communicate with hosts over different protocols
- [pkg/etcd](pkg/etcd) - allows to create and orchestrate etcd clusters running as containers
- [pkg/kubelet](pkg/kubelet) - allows to create and orchestrate kubelet pools running as containers
- [pkg/controlplane](pkg/controlplane) - allows to create and orchestrate Kubernetes static control plane containers
- [pkg/apiloadbalancer](pkg/apiloadbalancer) - allows to create and orchestrate loadbalancer containers for Kubernetes API server

Currently, the main strategy this library use is to talk to container runtime on remote hosts using SSH. However, as the library tries to separate data processing
from operational logic, other modes can be implemented easily in the future, for example:
- Binary packed in Docker container, which runs and evaluates local config
- Creating local configs per host to feed the binary mentioned above
- Kubernetes operator, which would be capable of SSH-ing into the cluster nodes and performing updates
- `kubectl exec` as a transport protocol

## Features

* Minimal host requirements - Use SSH connection forwarding for talking directly to the container runtime on remote machines for managing static containers and configuration files.
* Independent management of etcd, kubelets, static control plane and self-hosted components.
* All self-hosted control plane components managed using helm (e.g CoreDNS).
* 1st class support for Terraform provider for automation.
* Others:
  * etcd, kubelet and static control plane running as containers.
  * Self-hosted control plane.
  * Supported container runtimes:
    * Docker
  * Configuration via YAML or via Terraform.
  * Deployment using CLI tools or via Terraform.
  * HAProxy for load-balancing and failover between API servers.

## Requirements

Using this library has minimal target host (where containers will run) requirements:

- configured one of the supported container runtimes (currently only Docker)
- when deploying to remote hosts, SSH access with the user allowed to create containers (e.g. when using `Docker` as a container runtime,
  user must be part of `docker` group)

Direct `root` access (via SSH login or with e.g. `sudo`) on the target hosts is NOT required, as all configuration files are managed using temporary configuration containers.

No Public DNS or any other public discovery service is required for getting cluster up and running either.

[Managing certificates](#managing-certificates) section recommends using Terraform for generating TLS certificates for the cluster.
For that, the Terraform binary is required. It can be downloaded from the [official website](https://www.terraform.io/downloads.html).

## User tools

### CLI binaries

Currently, following binaries are implemented:
- [cmd/container-runner](cmd/container-runner) - low-level tool, which allows to create any kind of containers
- [cmd/etcd-cluster](cmd/etcd-cluster) - allows to create and manage etcd clusters
- [cmd/api-loadbalancers](cmd/api-loadbalancers) - allows to create and manage Kubernetes API loadbalancers
- [cmd/controlplane](cmd/controlplane) - allows to create and manage Kubernetes static control plane
- [cmd/kubelet-pool](cmd/kubelet-pool) - allows to create and manage kubelet pools
- [cmd/helm-release](cmd/helm-release) - allows to create and update helm 3 releases

All mentioned CLI binaries currently takes `config.yaml` file from working directory as an input
and produces `state.yaml` file when they finish working. `state.yaml` file contains information about created containers and it shouldn't
be modified manually. On subsequent runs, `state.yaml` file is read as well and it's used to track and update created containers.

Both `state.yaml` and `config.yaml` should be kept secure, as they will contain TLS private keys for all certificates used by the cluster.

### Terraform provider

In addition to CLI binaries, there is Terraform provider `terraform-provider-flexkube` available in [cmd/terraform-provider-flexkube](cmd/terraform-provider-flexkube).
This provider allows to create all resources mentioned above using Terraform.

Current implementation is very minimal, as each resource simply takes `config.yaml` file content using `config` parameter and it stores `state.yaml`
file content in `state` computed parameter, which is stored in the Terraform state. However, it already allows fully automated deployment of
entire Kubernetes cluster.

## Example usage

In [examples](examples) directory, you can find example configuration files for each CLI tool. In [examples/terraform-provider-flexkube](examples/terraform-provider-flexkube) you can find Terraform snippets, which automates creating valid configuration file and creating the resource itself.

All examples are currently configured to create containers on the local file-system. Please note, that creating containers will create configuration files on the root file-system, so it may override some of your files! Testing in the isolated environment is recommended. For example creating etcd member, will override files mentioned [here](pkg/etcd/member.go#L42).

To quickly try out each `config.yaml` file, you can use following command:
```
go run github.com/flexkube/libflexkube/cmd/<name of the tool>
```

For example, if you want to create simple container, run following command in [examples/container-runner](examples/container-runner) directory:
```
go run github.com/flexkube/libflexkube/cmd/container-runner
```

### Deploying over SSH

If you want to deploy to the remote host over SSH, you can change following configuration fragment:

```
host:
  direct: {}
```

to SSH configuration:

```
host:
  ssh:
    address: "<remote host>"
    port: 2222
    user: "core"
    privateKey: |
      <SSH private key to use>
```

If you deploy to multiple hosts over SSH, you can also define default SSH configuration, by placing `ssh` block on the top level of the configuration. For example, if you always want to use port 2222:
```
ssh:
  port: 2222
```

All those defaults can be overridden per instance of your container:
```
kubelets:
- host:
    ssh:
      address: "<remote host>"
      port: 3333
```

You can also authenticate with SSH using password:
```
ssh:
  password: "foo"
```

SSH agent authentication is currently NOT supported.

## Supported container runtimes

Currently only Docker is supported as a container runtime. Support for more container runtimes should be added in the future.

Each container runtime must implement [runtime](pkg/container/runtime/runtime.go) interface.

## Supported transport protocols

Currently, there are 2 transport protocols, which are supported:
- `direct` - which use local filesystem (for accessing UNIX sockets) and local network (for TCP connections) when accessing container runtimes. Note that this method also allows to communicate with remote daemons, if they are reachable over network. However, the connection may not be encrypted.
- `ssh` - which allows talking with container runtimes using SSH forwarding

In the future support for more protocols, like `winrm` might be implemented.

Each transport protocol must implement [transport](pkg/host/transport/transport.go) interface.

## Managing certificates

All deployments require X.509 certificates to secure the communication between components with TLS.

Following Terraform modules are the recommended way to manage those certificates:
- [terraform-root-pki](https://github.com/flexkube/terraform-root-pki) - generates root CA certificate, which should be used to sign intermediate CA certificates (etcd CA, Kubernetes CA etc).
- [terraform-etcd-pki](https://github.com/flexkube/terraform-etcd-pki) - generates etcd CA certificate, peer certificates, server certificates and client certificates
- [terraform-kubernetes-pki](https://github.com/flexkube/terraform-kubernetes-pki) - generates Kubernetes CA and all other required certificates for functional Kubernetes cluster

In the future go package might be added to manage them, to avoid having Terraform as a dependency.

## Getting started

To get stated, see [docs/GETTING-STARTED.md](docs/GETTING-STARTED.md)

## Current known issues and limitations

Currently, there are several things, which are either missing or broken. Here is the list of known problems:
- gracefully replacing CA certificates (if private key does not change, it should work, but has not been tested)
- no checkpointer for pods/apiserver. If static kube-apiserver container is stopped and node reboots, single node cluster will not come back.

And features, which are not yet implemented:
- network policies for kube-system namespace
- caching port forwarding
- using SSH agent for authentication
- bastion host(s) support for SSH
- parallel deployments across hosts
- removal of config files, created data and containers
- automatic shutdown/start of bootstrap control plane

## Testing

To see how to run tests during development, see [TESTING.md](docs/TESTING.md).

## Helm charts

All self-hosted control-plane deployments and CNI plugins are managed using [Helm](https://helm.sh/). All used charts are available via `https://flexkube.github.io/charts/` charts repository.

The repository is hosted using GitHub Pages and and it's content can be found in this [charts](https://github.com/flexkube/charts) repository.

## Contributing

All contributions to this project are welcome. If it does not satisfy your needs, feel free to raise an issue about it or implement the support yourself and create a pull request with the patch, so we can all benefit from it.

If you just want to help the project grow and mature, there are many TODOs spread across the code, which should be addresses sooner or later.

## Status of the project

At the time of writing, this project is in active development state and it is not suitable for production use. Breaking changes
might be introduced at any point in time, for both library interface and for existing deployments.

Currently, there is no good documentation describing how to configure and use implemented tools. Digging into the source code is highly recommended.
With help of error messages as trace points, the code should be clear enough to figure out the right configuration.

More examples of use will be added in the future.
