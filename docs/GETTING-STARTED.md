# Getting started

This document describes how to get started with basic functionality of the project.

## Table of contents

- [Deploying etcd cluster](#deploying-etcd-cluster)
- [Deploying API Load Balancer (Optional)](#deploying-api-load-balancer--optional-)
- [Bootstrap control plane](#bootstrap-control-plane)
- [Installing self-hosted Kubernetes Control Plane](#installing-self-hosted-kubernetes-control-plane)
- [Deploying kubelet pools](#deploying-kubelet-pools)
- [Shutting down static control plane](#shutting-down-static-control-plane)

## Deploying etcd cluster

First thing required for Kubernetes cluster is etcd cluster. It can be deployed using [cmd/etcd-cluster](cmd/etcd-cluster) binary. Please refer to [pkg/etcd/cluster.go#L18](pkg/etcd/cluster.go#L18) and [pkg/etcd/member.go#L15](pkg/etcd/member.go#L15) for required configuration.

## Deploying API Load Balancer (Optional)

If you plan to have more than one control plane node, it is recommended to use the API Load Balancer, which will take care of HA on the client side, as currently `Kubelet` is not able to handle multiple API servers. It can be deployed using [cmd/api-loadbalancers](cmd/api-loadbalancers).

Configuration reference:
- [pkg/apiloadbalancer/api-loadbalancers.go#L16](pkg/apiloadbalancer/api-loadbalancers.go#L16)
- [pkg/apiloadbalancer/api-loadbalancer.go#L16](pkg/apiloadbalancer/api-loadbalancer.go#L16)

## Bootstrap control plane

Kubernetes static Control Plane is required to kick off the self-hosted control plane. It consist of minimal set of processes:
- kube-apiserver
- kube-controller-manager
- kube-scheduler

All those 3 containers can be created using [cmd/controlplane](cmd/controlplane) binary. The available parameters are described here: [pkg/controlplane/controlplane.go#L18](pkg/controlplane/controlplane.go#L18).

## Installing self-hosted Kubernetes Control Plane

Once bootstrap (static) control plane is running and functional, self-hosted version of it should be installed on top of that, to ensure better survivability and graceful updates.

Recommended way of doing that is using [kube-apiserver-helm-chart](https://github.com/flexkube/kube-apiserver-helm-chart) and [kubernetes-helm-chart](https://github.com/flexkube/kubernetes-helm-chart).

There are 2 helm charts for the controlplane, as [version-skew-policy/#supported-component-upgrade-order](https://kubernetes.io/docs/setup/release/version-skew-policy/#supported-component-upgrade-order) recommends, that `kube-apiserver` should be upgraded first, then remaining components of the control plane and helm currently does not support such ordering.

The helm charts can be installed using one of the following methods:
- using `helm` CLI (while helm 2.x could somehow work, we strongly recomment using helm 3.x, as this one does not require Tiller process to be running, so it can be used on static control plane straight away)
- using [cmd/helm-release](cmd/helm-release), which gives minimal interface to create helm 3 release on the cluster
- using `flexkube_helm_release` Terraform resource (as at the time of writing, [terraform-provider-helm](https://github.com/terraform-providers/terraform-provider-helm) does not support helm 3 yet. Upstream issue: https://github.com/terraform-providers/terraform-provider-helm/issues/299)

## Deploying kubelet pools

Even though the self-hosted control plane can be installed, it won't be running without any nodes registered in the cluster. In order to get nodes in the cluster, `kubelet`s needs to be deployed.

This can be done using [cmd/kubelet-pool](cmd/kubelet-pool), which deploys kubelet containers with the same configuration to multiple nodes. If you need kubelets with different configurations, please create multiple pools.

The configuration reference can be found in [pkg/kubelet/pool.go#L18](pkg/kubelet/pool.go#L18).

NOTE: kubelet pool have `serverTLSBootstrap: true` option enabled, so their serving certificates (for HTTPS communication coming from from kube-apiserver) will be requested from the cluster. Currently, such certificates are not automatically approved, so it is recommended to use [kubelet-rubber-stamp](https://github.com/kontena/kubelet-rubber-stamp) to automate approval process. It can be deployed using [kubelet-rubber-stamp](https://github.com/flexkube/kubelet-rubber-stamp-helm-chart) helm chart.

## Shutting down static control plane

Once the charts are installed and their pods are running, it is recommended to shut down static control plane, to prevent relying on it, as it won't receive the updates when helm charts configuration is updated.

This can be done in 2 ways:
- By running `docker stop kube-apiserver kube-scheduler kube-controller-manager` on the bootstrap host.
- By adding `shutdown: true` to `controlplane` resource and re-applying it. Please note that this will remove static containers as well.
