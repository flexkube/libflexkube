# Docs

This directory contains development documentation of libflexkube. For Flexkube documentation, visit [flexkube.github.io](https://flexkube.github.io).

## Table of contents

- [Testing](#testing)
  - [Unit tests](#unit-tests)
  - [Integration tests](#integration-tests)
  - [E2E tests](#e2e-tests)
  - [Conformance tests](#conformance-tests)
  - [Local tests](#local-tests)
    * [Target host](#target-host)
      + [VirtualBox](#virtualbox)
      + [Libvirt](#libvirt)
    * [Requirements](#requirements)
    * [Helm charts development](#helm-charts-development)
      + [Via Helm CLI](#via-helm-cli)
      + [Via Terraform](#via-terraform)
    * [Terraform modules development](#terraform-modules-development)
  - [Cleaning up](#cleaning-up)

## Testing

This section describes how to run various tests while developing the project.

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
To make running such tests easier, there is a `Vagrantfile` available, which spawns [Flatcar Container Linux](https://www.flatcar-linux.org/) virtual machine, where all further tests can be executed. Currently the only tested provider
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

This command will create testing virtual machine and run `TestE2E()` test from `e2e` directory to create Kubernetes cluster. At the end of the tests, `kubeconfig` file with admin access to the cluster will be copied to the `e2e` directory, which allows further inspection.

If you don't have `kubectl` available on host, following command can be executed to spawn shell in E2E container on virtual machine, which contains additional tools like `kubectl` or `helm` binaries and comes with `kubeconfig` predefined, to ease up testing:
```sh
make vagrant-e2e-shell
```

If you just want to run E2E tests and clean everything up afterwards, run the following command:
```sh
make vagrant-e2e
```

### Conformance tests

To run conformance tests in the environment provided by `Vagrantfile`, run the following command:
```
make vagrant-conformance
```

The command will deploy E2E environment and then run conformance tests in there.

The test should take a bit more than an hour to finish.

By default, after scheduling the conformance tests, the command will start showing the logs of the tests. One can then use CTRL-C to stop showing the logs, as tests will be running in the background and the command is idempotent.

Once tests are complete, the command should will the test results and archive file with the report will be copied into project's root directory, which can be then submitted to [k8s-conformance](https://github.com/cncf/k8s-conformance) repository.

### Local tests

For testing standalone resources, e.g. just `etcd-cluster`, [local-testing](./local-testing) directory can be used, which will use the code from [e2e](./e2e) directory to create a cluster and then will dump all configuration and state files to separate directories, when tools from [cmd](./cmd) directory can be used directly. That allows to skip many sync steps, which speeds up the overall process, making development easier.

#### Target host

##### VirtualBox

By default, local testing is configured to deploy to virtual machine managed by [Vagrantfile](./Vagrantfile), which can be brought up using the following command:
```sh
make vagrant-up
```

However, if you like to test on some other machine, you can override the following parameters, by creating `local-testing/variables.auto.tfvars` file:
- `ssh_private_key_path` - To provide your own private SSH key.
- `node_internal_ip` - This should be set to your target host IP, which will be used by cluster components.
- `node_ssh_port` - Target host SSH port.
- `node_address` - Target host SSH address and where `kube-apiserver` will be listening for client requests.

##### Libvirt

In addition to VirtualBox, also KVM with libvirt can be used to spawn local environment. The environment is managed with Terraform, so Terraform binary is required. To set it up, use the following command:
```sh
make libvirt-apply
```

The command will download Flatcar QEMU image and required Terraform providers. On subsequent runs, those steps will be skipped. After that, libvirt network, pools and machines will be created.

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

Charts can also be tested using Terraform. This can be done by creating `local-testing/test-config.yaml` file, with following example content:
```yaml
charts:
  kubeAPIServer:
    source: <local path with cloned kube-apiserver chart>
    version: ">0.0.0-0"
```

Then run the following command to deploy updated chart:
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
