# etcd lifecycle management library

This project helps deploying and maintaining lifecycle of etcd deployment. Because of modular
architecture, it can be used in various environments and with different configurations and backends.

## Frontends

Currently, the main purpose of this library is to serve for Terraform Provider, however there are
various frontends, which could be implemented, for example:
- CLI tool
- containerized daemon
- direct library use

## Backends

As etcd deployment is stateful, the state about machines needs to be stored somewhere (for example to track
which nodes should be removed from the cluster.

The state can be stored in following locations:
- Amazon S3 or compliant storage implementation (not implemented)
- Google GCS (not implemented)
- local file (not implemented)

The state will be also used to lock on cluster operations, to make sure only single process modifies
cluster at a time.

## Modes of operation

Because of shared state storage, various frontends can be combined and used together with different
purposes. The main ones are:
- controller - reads the cluster state, and schedules changes, but not applying changes (this will
  be done by for example daemon running on the cluster).
- monitor - only reads cluster state and checks if it's up to date. Can be used as monitorig/health check tool.
- executor - periodically reads desired cluster state and executes it.

## Difference between similar projects

There are many other project, which purpose is to deploy etcd. This section covers briefly what
is the difference between this and those projects.

- etcdadm
- etcd-operator
- rke
