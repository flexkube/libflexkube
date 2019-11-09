# libflexkube: Go library for deploying kubernetes

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

On executig host, the user-writable `/tmp` directory is required. This is where the forwarded UNIX sockets are stored.
In the future this requirement might be removed, if we implement using virtual sockets.

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

All those defaults can be overriden per instance of your container:
```
kubelets:
- host:
    ssh:
      address: "<remote host>"
      port: 3333
```

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
- [terraform-root-pki](https://github.com/invidian/terraform-root-pki) - generates root CA certificate, which should be used to sign intermediate CA certificates (etcd CA, Kubernetes CA etc).
- [terraform-etcd-pki](https://github.com/invidian/terraform-etcd-pki) - generates etcd CA certificate, peer certificates and client certificates
- [terraform-kubernetes-pki](https://github.com/invidian/terraform-kubernetes-pki) - generates Kubernetes CA and all other required certificates for functional Kubernetes cluster

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

Recommended way of doing that is using [kubernetes-helm-chart](https://github.com/invidian/kubernetes-helm-chart).

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

## Contributing

All contributions to this project are welcome. If it does not satisfy your needs, feel free to raise an issue about it or implement the support yourself and create a pull request with the patch, so we can all benefit from it.

If you just want to help the project grow and mature, there are many TODOs spread across the code, which should be addresses sooner or later.
