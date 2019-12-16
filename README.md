# libflexkube: Go library for deploying kubernetes

[![Build Status](https://travis-ci.org/flexkube/libflexkube.svg?branch=master)](https://travis-ci.org/flexkube/libflexkube) [![Maintainability](https://api.codeclimate.com/v1/badges/5840c3fe0a9bc77aef08/maintainability)](https://codeclimate.com/github/flexkube/libflexkube/maintainability) [![Test Coverage](https://api.codeclimate.com/v1/badges/5840c3fe0a9bc77aef08/test_coverage)](https://codeclimate.com/github/flexkube/libflexkube/test_coverage) [![codecov](https://codecov.io/gh/flexkube/libflexkube/branch/master/graph/badge.svg)](https://codecov.io/gh/flexkube/libflexkube)

## Status of the project

At the time of writing, this project is in active development state and it is not suitable for production use. Breaking changes
might be introduced at any point in time, for both library interface and for existing deployments.

Currently, there is no good documentation describing how to configure and use implemented tools. Digging into the source code is highly recommended.
With help of error messages as trace points, the code should be clear enough to figure out the right configuration.

More examples of use will be added in the future.

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

## Requirements

Using this library has minimal target host (where containers will run) requirements:

- configured one of the supported container runtimes (currently only Docker)
- when deploying to remote hosts, SSH access with the user allowed to create containers (e.g. when using `Docker` as a container runtime,
  user must be part of `docker` group)

`root` access on the target hosts is NOT required, as all configuration files are managed using temporary configuration containers.

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
- [terraform-etcd-pki](https://github.com/flexkube/terraform-etcd-pki) - generates etcd CA certificate, peer certificates and client certificates
- [terraform-kubernetes-pki](https://github.com/flexkube/terraform-kubernetes-pki) - generates Kubernetes CA and all other required certificates for functional Kubernetes cluster

In the future go package might be added to manage them, to avoid having Terraform as a dependency.

## Getting started

### Deploying etcd cluster

First thing required for Kubernetes cluster is etcd cluster. It can be deployed using [cmd/etcd-cluster](cmd/etcd-cluster) binary. Please refer to [pkg/etcd/cluster.go#L18](pkg/etcd/cluster.go#L18) and [pkg/etcd/member.go#L15](pkg/etcd/member.go#L15) for required configuration.

### Deploying API Load Balancer (Optional)

If you plan to have more than one control plane node, it is recommended to use the API Load Balancer, which will take care of HA on the client side, as currently `Kubelet` is not able to handle multiple API servers. It can be deployed using [cmd/api-loadbalancers](cmd/api-loadbalancers).

Configuration reference:
- [pkg/apiloadbalancer/api-loadbalancers.go#L16](pkg/apiloadbalancer/api-loadbalancers.go#L16)
- [pkg/apiloadbalancer/api-loadbalancer.go#L16](pkg/apiloadbalancer/api-loadbalancer.go#L16)

### Bootstrap control plane

Kubernetes static Control Plane is required to kick off the self-hosted control plane. It consist of minimal set of processes:
- kube-apiserver
- kube-controller-manager
- kube-scheduler

All those 3 containers can be created using [cmd/controlplane](cmd/controlplane) binary. The available parameters are described here: [pkg/controlplane/controlplane.go#L18](pkg/controlplane/controlplane.go#L18).

### Installing self-hosted Kubernetes Control Plane

Once bootstrap (static) control plane is running and functional, self-hosted version of it should be installed on top of that, to ensure better survivability and graceful updates.

Recommended way of doing that is using [kubernetes-helm-chart](https://github.com/flexkube/kubernetes-helm-chart).

The helm chart can be installed using one of the following methods:
- using `helm` CLI (while helm 2.x could somehow work, we strongly recomment using helm 3.x, as this one does not require Tiller process to be running, so it can be used on static control plane straight away)
- uning [cmd/helm-release](cmd/helm-release), which gives minimal interface to create helm 3 release on the cluster
- using `flexkube_helm_release` Terraform resource (as at the time of writing, [terraform-provider-helm](https://github.com/terraform-providers/terraform-provider-helm) does not support helm 3 yet. Upstream issue: https://github.com/terraform-providers/terraform-provider-helm/issues/299)

Once the chart is installed and it's pods are running, it is recommended to shut down static control plane, to prevent relying on it, as it won't receive the updates when helm chart configuration is updated.

Currently, this needs to be done manually using `docker stop` command.

### Deploying kubelet pools

Even though the self-hosted control plane can be installed, it won't be running without any nodes registered in the cluster. In order to get nodes in the cluster, `kubelet`s needs to be deployed.

This can be done using [cmd/kubelet-pool](cmd/kubelet-pool), which deploys kubelet containers with the same configuration to multiple nodes. If you need kubelets with different configurations, please create multiple pools.

The configuration reference can be found in [pkg/kubelet/pool.go#L18](pkg/kubelet/pool.go#L18).

NOTE: kubelet pool have `serverTLSBootstrap: true` option enabled, so their serving certificates (for HTTPS communication coming from from kube-apiserver) will be requested from the cluster. Currently, such certificates are not automatically approved, so it is recommended to use [kubelet-rubber-stamp](https://github.com/kontena/kubelet-rubber-stamp) to automate approval process. It can be deployed using [kubelet-rubber-stamp](https://github.com/flexkube/kubelet-rubber-stamp-helm-chart) helm chart.

## Current known issues and limitations

Currently, there are several things, which are either missing or broken. Here is the list of known problems:
- fetching logs from pods using `kubectl logs`
- network plug-ins are not configurable (currently `kubenet` is hardcoded)
- gracefully replacing CA certificates (if private key does not change, it should work, but has not been tested)
- adding/removing etcd members
- surviving reboot not tested

And features, which are not yet implemented:
- TLS encryption between etcd and kubernetes API server
- pod security policies for for control plane pods
- network policies for kube-system namespace
- caching port forwarding
- batching config file updates
- using SSH agent for authentication
- bastion host(s) support for SSH
- paralllel deployments across hosts
- showing diff to the user (planning what will be done)
- removal of config files, created data and containers
- automatic shutdown/start of bootstrap control plane
- taints and tolerations for control plane
- role labels for kubelets

## Testing

### Unit tests

Unit tests can be executed by running the following command:
```sh
make test
```

If you want to only run tests for specific package, you can add `GO_PACKAGES` variable to the `make` command. For example, to run tests only in `pkg/host/transport/ssh/` directory, execute the following command:
```sh
make test GO_PACKAGES=./pkg/host/transport/ssh/...
```

### Integration tests

As this library interacts a lot with other systems, in addition to unit tests, there are also integration tests defined, which has extra environmental requirements, like Docker daemon running or SSH server available.
To make running such tests easier, there is a `Vagrantfile`, which spawns [Flatcar Container Linux](https://www.flatcar-linux.org/) virtual machine, where all further tests can be executed. Currently the only tested provider
for Vagrant is VirtualBox.

In order to run integration tests in virtual environment, execute the following command:
```sh
make vagrant-integration
```

This will spawn the virtual machine, copy source code into it, build Docker container required for testing and run the integration tests. This target is idempotent and when run multiple times, it will use cached test results and Docker images, so subsequent runs should be much faster, which allows to modify tests on host machines and re-run them in testing environment.

This target also allows you to override `GO_PACKAGES` to only run specific integration tests.

In the repository, integration tests files has `*_integration_test.go` suffix and they use `integration` Go build tag to be excluded from regular testing. All tests, which has some environment requirements should be specified as integration tests, to keep unit tests minimal and always runnable, for example in the CI environment.

To debug issues with integration tests, following command can be executed to spawn the shell in Docker container running on testing virtual machine:
```sh
make vagrant-integration-shell
```

### E2E tests

In addition to integration tests, this repository has also defined E2E tests, which tests overrall functionality of the library. As integration tests, they are also executed inside the virtual machine.

E2E tests can be executed using following command:
```sh
make vagrant-e2e-run
```

This command will create testing virtual machine, compile Flexkube Terraform provider inside it and then use Terraform to create Kubernetes cluster. At the end of the tests, `kubeconfig` file with admin access to the cluster will be copied to the project's root directory, which allows further inspection.

If you don't have `kubectl` available on host, following command can be executed to spawn shell in E2E container on virtual machine, which contains additional tools like `kubectl` or `helm` binaries and comes with `kubeconfig` predefined, to ease up testing:
```sh
make vagrant-e2e-shell
```

If you just want to run E2E tests and clean everything up afterwards, run the following command:
```sh
make vagrant-e2e
```

### Local tests

For testing standalone resources, e.g. just `etcd-cluster`, [local-testing](./local-testing) directory can be used, which will use the code from [e2e](./e2e) directory to create a cluster and then will dump all configuration and state files to separate directories, when tools from [cmd](./cmd) directory can be used directly. That allows to skip many sync steps, which speeds up the overall process, making development easier.

#### Target host

By default, local testing is configured to deploy to virtual machine managed by [Vagrantfile](./Vagrantfile), which can be brought up using the following command:
```sh
make vagrant-up
```

However, if you like to test on some other machine, you can override the following parameters, by creating `local-testing/variables.auto.tfvars` file:
- `ssh_private_key_path` - To provide your own private SSH key.
- `node_internal_ip` - This should be set to your target host IP, which will be used by cluster components.
- `node_ssh_port` - Target host SSH port.
- `node_address` - Target host SSH address and where `kube-apiserver` will be listening for client requests.

#### Requirements

Local testing runs compilation and Terraform locally, so both `go` and `terraform` binaries needs to be available.

If testing using virtual machine, `vagrant` is required to set it up.

If testing with Helm CLI, `helm` binary is needed.

#### Helm charts development

In addition to resources config files, also `values.yaml` files for all chart releases are dumped to the `local-testing/values` directory. This allows to also make developing the helm charts easier.

##### Via Helm CLI

To update the chart using `helm` binary, run following command:
```sh
helm upgrade --install -n kube-system -f ./local-testing/values/kubernetes.yaml kubernetes <path to cloned chart source>
```

##### Via Terraform

Charts can also be tested using Terraform. This can be done by creating `local-testing/variables.auto.tfvars` file, with following example content:
```
kube_apiserver_helm_chart_source = "<local path with cloned kube-apiserver chart>"
```

Then run the following command to deploy updated chart:
```sh
make test-local-apply
```

#### Terraform modules development

Local testing is also handy for testing changes to Terraform modules like [terraform-etcd-pki](https://github.com/flexkube/terraform-etcd-pki).

To test changes, modify [local-testing/main.tf](./local-testing/main.tf) file and change the source of the desired module to point to your copy. Then run the following command:
```sh
make test-local-apply
```

### Cleaning up

After finished testing, following command can be executed to clean up testing virtual machine:
```sh
make vagrant-destroy
```

If you want to also remove all artifacts from the repository, like built binaries, coverage files etc, run the following command:
```sh
make clean
```

## Contributing

All contributions to this project are welcome. If it does not satisfy your needs, feel free to raise an issue about it or implement the support yourself and create a pull request with the patch, so we can all benefit from it.

If you just want to help the project grow and mature, there are many TODOs spread across the code, which should be addresses sooner or later.
