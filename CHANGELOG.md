# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.9.0] - 2022-09-13
### Changed
- Default Kubernetes version is now v1.25.0.
- Default etcd version is now v3.5.4.
- Default HAProxy version is now 2.6.5.
- Default Calico CNI version is now v3.24.1.
- Binaries are built using Go 1.19.

### Removed
- As part of Kubernetes v1.25 upgrade, support for PodSecurityPolicies has been removed.

## [0.8.0] - 2022-05-05
### Changed
- When `flexkube` binary is installed using `go install`, it will print version information
pulled from Go modules.
- Default Kubernetes version is now v1.24.0.
- Default etcd version is now v3.5.3.
- Default HAProxy version is now 2.5.6.
- Binaries are built using Go 1.18.

### Removed
- As part of Kubernetes v1.24 upgrade support for selecting network plugin in kubelet has been removed.

## [0.7.0] - 2021-09-02
### Added
- Certificates and private keys in configs are now parsed as part of validation process.
- e2e tests has now container runtime, nodes CIDR, and kubelet extra args configurable.
- e2e tests are now run as part of CI process.
- CI process now covers Dockerfiles, changelog formatting, go mod tidyness, Vagrantfile and Terraform code.
- Custom code style checks using semgrep.

### Changed
- Default Kubernetes version is now v1.22.1.
- Default etcd version is now v3.5.0.
- Default HAProxy version is now v2.4.3.
- Error messages has been improved across all codebase.
- CI and Docker images now use Go 1.17.
- Binaries are now build with local paths stripped (-trimpath flag).
- e2e and local tests now use containerd as container runtime.
- Sonobuoy version used in conformance tests has been updated to latest version v0.53.2.
- golangci-lint version v1.42.0 is now used.

## [0.6.0] - 2021-05-24
### Changed
- `etcd.Member` struct has been renamed to `etcd.MemberConfig` and `etcd.Member` is now an
interface due to internal refactoring.
- Updated Go dependencies to latest versions.
- Default Kubernetes version is now v1.21.1.
- Default HAProxy version is now 2.3.10.
- Default etcd version is now v3.4.16.
- Vagrant is now using Docker again instead of containerd as container runtime due to some conformance tests failing.

### Fixed
- Missing newline in removing container configuration log message.
- Running e2e tests using Vagrant when having local e2e test configuration.

## [0.5.1] - 2021-02-19
### Changed
- Default Kubernetes version is now v1.20.3.
- Default HAProxy version is now 2.3.5.
- Switched CI from Travis to GitHub Actions. This results in faster updates to PR statuses.

### Fixed
- Logic for running conformance tests, they should be now more robust.

## [0.5.0] - 2020-12-11
### Added
- It is now possible to pass extra flags to kubelet container via 'extraFlags' field in kubelet and pool configuration.
This combined with extra mounts allows to switch kubelet to use containerd as a container runtime instead of now
deprecated Docker.

### Changed
- Improved handling situations when node hosting a container is gone. Previously trying to apply configuration
in such situation would result in an error, which forced user to either get a host back or to manually
modify the state to get rid of old container.
- Due to Kubernetes version update, controlplane parameter for kube-apiserver 'serviceAccountPublicKey' has been
replaced with 'serviceAccountPrivateKey', as kube-apiserver now requires private key and public key can be
derived from private one. For users using PKI integration, there is no expected changes.
- Default Kubernetes version is now v1.20.0.
- Default HAProxy version is now 2.3.2.
- Default etcd version is now v3.4.14.
- Default Calico version is now v3.17.1.
- Go version used for building the binaries is now 1.15.6.
- Various e2e tests improvements.

## [0.4.3] - 2020-09-20
### Added
- Changing `IPAddresses` field for PKI certificates and running `Generate()` will now properly
re-generate the certificate to align the field with the configuration. Additionally, it is now
easier to add more rules for certificate re-generation, for example based on expiry time. This
might be done in further releases.
- etcd cluster and members may have now additional mounts configured. This is a ground work for allowing
etcd to listen on UNIX sockets and for generic resource customization.
- In case when Kubernetes API returns etcd-related error, changing Helm release will now retry the operation,
as in most cases it works on 2nd attempt. If 3 consecutive errors occur, error is returned. This will make
adding and removing controller nodes more robust.

### Changed
- All Helm charts used has been updated to the latest versions.
- Default HAProxy version is now `2.2.3`.
- Default Kubernetes version is now `v1.19.2`.
- Generated PKI certificates will now only include generated values instead of all values, as
some of them are inherited from other fields and including them there breaks updating via
inherited fields.
- Generation of etcd certificates via PKI is now improved. Now all changes to `Peers` and `Servers`
fields are properly propagated and all properties are properly inherited.
- Maps in `PKI.Etcd` are now only initialized if there are some certificates to be stored.
- etcd now uses explicit rules for validating certificates and private key fields, so error messages
will be better if any of those fields is malformed.

### Fixed
- etcd certificates generated using `PKI` now always include `127.0.0.1` server address, to make
sure that adding etcd members via SSH port forwarding works as expected. This broke adding/removing
etcd members if PKI integration was used.

## [0.4.2] - 2020-09-16
### Fixed
- Release [0.4.1](https://github.com/flexkube/libflexkube/compare/v0.4.0...v0.4.1) has been tagged wrongly and ended up not including `env` key support for
containers. This is now fixed.

## [0.4.1] - 2020-09-15
### Added
- flexkube: Added `template` subcommand which can be fed with Go template which will have access
to CLI resource configuration and state, which allows generating Helm values.yaml files for
self-hosted controlplane charts.
- container: Added support for defining environment variables for containers using `env` key.

### Changed
- e2e: Use Go test framework rather than Terraform to create a cluster.
- Updated `golangci-lint` to version `v1.31.0`.
- Default Kubernetes version is now `v1.19.1`.

### Removed
- Terraform provider code is now removed and lives in
[flexkube/terraform-provider-flexkube](https://github.com/flexkube/terraform-provider-flexkube)
repository.

## [0.4.0] - 2020-08-31
### Changed
- e2e: Updated used sonobuoy version to v0.19.0.
- e2e/local-testing: use Terraform 0.13.
- Default Kubernetes version is now v1.19.0.
- As upstream Kubernetes deprecated hyperkube image, now each controlplane component
use individual images. As upstream does not publish kubelet images yet, new default kubelet image
is build from [kubelet](https://github.com/flexkube/kubelet) repository and available for pulling
from `quay.io/flexkube/kubelet` registry.
- controlplane: static kube-apiserver now runs on host network and with `--permit-port-sharing=true`
flag set to make use of binding with SO_REUSEPORT option, which eliminates the need of bootstrap
HAProxy and HAProxy container on self-hosted kube-apiserver pods.
- e2e: use Helm v3.3.0.
- Updated Go dependencies to latest versions.

### Removed
- e2e: Remove bootstrap API Load Balancer - it is no longer needed as since Kubernetes v1.19.0,
kube-apiserver is able to bind with SO_REUSEPORT, if `--permit-port-sharing=true` flag is set.

## [0.3.3] - 2020-08-29
### Changed
- Updated Calico to v3.16.0.

### Fixed
- Fixed kubelet applying process panicking, when `WaitForNodeReady` is `true` and `AdminConfig`
is not specified. Now `WaitForNodeReady` requires `AdminConfig`, as waiting action is executed
on the client side, similar to applying privileged labels to the node.

## [0.3.2] - 2020-08-28
### Changed
- Default Kubernetes version is now v1.18.8.
- Default HAProxy version is now v2.2.2.
- Default etcd version is now v3.4.13.
- linter: Updated golangci-lint to v1.30.0.
- conformance: Dpdated sonobuoy version to v0.18.5.
- e2e: Pinned Terraform version to allow running conformance tests on old versions
in the future.
- e2e: pinned Kubernetes version and Helm charts versions to allow running conformance
tests on old version in the future.
- Updated Golang version used on CI to 1.15.

### Fixed
- `Version` parameter is now respected when managing Helm releases.
- Helm release now exposes Helm's --wait option via `Wait` field.
- Improved reliability of running conformance tests in e2e environment.

## [0.3.1] - 2020-07-31
### Added
- `flexkube` CLI will now print colored diff when configuration changes are detected.
- `flexkube` CLI will now ask user for confirmation before deploying the resources, unless `--yes` flag is set.
- `flexkube` CLI now supports `--noop` flag, which allows only checking if the configuration is up to date, without triggering the deployment.
- `flexkube` CLI now supports `conatiners` sub-command for managing arbitrary groups of containers. This allows to also manage some extra containers not provided by `libflexkube`.
- `pkg/kubelet` now supports waiting until node gets into ready state, if `WaitForNodeReady` flag is set to `true`.
- `kube-apiserver` from static controlplane now use `--target-ram-mb` flag to limit memory usage of bootstrap controlplane.

### Changed
- New website with user documentation is now available at [flexkube.github.io](https://flexkube.github.io/). The documentation is not complete yet, but it's already better than existing documentation.
- `kube-proxy` and TLS bootstrapping rules are now installed using separate Helm Charts. This is because in case of managed cluster, those components must be installed on the target cluster, not on management cluster. It also allows specifying multiple bootstrap tokens, for example per kubelet pool.
- Improved the documentation of all Go packages.
- Updated Helm binary in `e2e` tests to `v3.2.3` and `sonobuoy` binary to `v0.18.4`.
- Updated all Go dependencies to latest versions.
- Updated default Kubernetes version to `1.18.6`.
- Updated default HAProxy version to `2.2.0`.
- Updated default etcd version to `3.4.10`.
- Mountpoints for containers are now created with `0700` permissions by default to increase security and satisfy etcd requirements. Existing users should make sure that `/var/lib/etcd/*` directories has `0700` permissions, otherwise etcd won't start after the upgrade.

### Fixed
- controlplane configuration won't be now validated, when `destroy: true` is specified. That allows removing entire configuration and running the deployment, which will then only validate the state of the deployment and remove all managed containers. This allows easy way of cleaning up when using `flexkube controlplane` command.
- All certificates generated by PKI has now `SubjectKeyID` set.
- `PeerCertAllowedCN` is now correctly used in `etcd` when it's explicitly defined, which should fix TLS connectivity issues in some setups.
- Fixed Helm release resource creating resources in the wrong namespace.
- `flexkube_helm_release` no longer leaks kubeconfig and values into plan, as they may contain sensitive information.

### Removed
- `containerrunner` binary is now replaced by `flexkube containers` subcommand.
- `helm-release` binary is now removed. Users are recommended to use official `helm` binary.

## [0.3.0] - 2020-05-24
### Added
- Added new `flexkube` CLI binary, which allows to manage multiple resources with the same configuration file. It replaces old `etcd-cluster`, `controlplane`, `api-loadbalancers`, `kubelet-pool` and `pki-generator` binaries.
- Added `PKI` resource, which allows generating all certificates required for cluster using Go API, as Terraform `flexkube_pki` resource or using `flexkube pki` command. This replaces [terraform-root-pki](https://github.com/flexkube/terraform-root-pki), [terraform-etcd-pki](https://github.com/flexkube/terraform-etcd-pki) and [terraform-kubernetes-pki](https://github.com/flexkube/terraform-kubernetes-pki) Terraform modules.
- Controlplane, etcd and kubelet-pool resources have now PKI resource integration with extra PKI field, so certificates no longer need to be generated externally and provided in configuration. This should simplify the use of CLI tools and Go API.
- SSH transport method now automatically integrates with `ssh-agent` if `SSH_AUTH_SOCK` environment variable is set. This allows using this transport method without any credentials configured.

### Changed
- Improved error messages when resource has no instances configured.
- Updated all dependencies to latest versions to fix installing using `go get`.
- Updated `sonobuoy` to `0.18.1`.
- State files are now created with `0600` permissions.
- Updated `golangci-lint` to `1.27.0`.
- Kubelet now use structured configuration instead of kubeconfig-like string field for bootstrap and administrator kubeconfig fields.
- `e2e` testing environment now use new PKI resource.
- Terraform provider unit tests no longer requires `tls` provider and all run in parallel, so they should be a bit faster to execute.
- Updated default `etcd` version to `3.4.9`.
- `VolumePluginDir` and `NetworkPlugin` fields now use default values for Kubelet and Controlplane resources, to minimize the default configuration required from the user.
- Release binaries now ship with stripped debug symbols, which makes them smaller.

### Fixed
- Constant diff in `containers-runner` and `flexkube_containers` resources caused by wrong JSON struct tags.
- When removing containers in `restarting` state, they will also be stopped before removing. Before, restarting containers requires manual stop to be removed.
- Bunch of typos.

### Removed
- Removed `etcd-cluster`, `controlplane`, `api-loadbalancers`, `kubelet-pool` and `pki-generator` binaries, replaced by `flexkube`.

## [0.2.2] - 2020-04-19
### Added
- It is now possible to configure extra mounts for kubelet container via extraMounts/extra_mount parameters
- etcd is now ready for enabling RBAC
- local-testing environment now generates script for enabling etcd RBAC

### Changed
- Default Kubernetes version is now 1.18.2
- Default HAProxy version is now 2.1.4
- Default etcd version is now 3.4.7
- Improved validation rules of controlplane. Now state from previous deployments will be validated as well.

### Fixed
- HAProxy now use HTTPS for probing kube-apiserver to avoid extensive logging of TLS handshake errors
- HAProxy configuration no longer generates warnings
- Fixed destroying flexkube_controlplane resource
- It is now possible to add and remove nodes in local-testing environment

## [0.2.1] - 2020-03-30
### Fixed
- libvirt worker nodes now use correct ignition config, not controller ones
- e2e/libvirt - reduce reserved RAM on worker nodes to 100Mi
- terraform: fix reporting inconsistent plan when config files changes
- adding and removing etcd members
- adding and removing controller nodes in e2e environment does not cause inconsistent plan anymore

## [0.2.0] - 2020-03-17
### Added
- Support for adding and removing etcd members (#28)
- libvirt as local testing environment (#34)
- Project logo and Certified Kubernetes logo (#36)
- Enabled NodeRestriction admission plugin (#35)
- Added support for specifying user and group when running containers (#57)
- Self-hosted and bootstrap kube-apiserver instances can now run in parallel, by adding a HAProxy load balancer
in front of them, which use SO_REUSEPORT socket option. This also allows to do graceful upgrades of self-hosted
kube-apiserver pod, as more than 1 instance can run in parallel on a single controller node. This prevents self-hosted
instance from crashing until bootstrap one is stopped. (#59)
- Show diff when applying changed from CLI tool (#65)
- Support for running mutationt tests
- Enabled PSP admission and added policies for all controlplane workloads
- Deploy metrics-server for local-testing and e2e environments
- hosts can now forward TCP connections

### Changed
- Updated golangci-lint to 1.23.8 (#31, #32, #68)
- Fixed all code smells reported by Code climate (#50)
- Re-enabled dupl and golint linters (#76, #69)
- Migrated Terraform resources to use native schema, show nice diffs to the users and trigger resource updates
if configuration or conditions changes.
- Terraform provider now can correctly destroy all the resources (#78, #82)
- Improved idempotency of both CLI tools and Terraform provider. Now if any action fails, all already build state
will be persisted, so once configuration or external conditions are fixed, user can proceed with the deployment (#42)
- Split README.md into smaller documents and added ToC for all of them (#87)
- Updated default Kubernetes version to 1.17.4
- Improved overall unit test coverage
- Improved quality of unit tests for some packages with mutation testing
- Updated default HAProxy version to 2.1.3
- Migrated Terraform provider to use terraform-plugin-sdk
- Bootstrap controlplane and API load balancers now run as unprivileged users
- Updated default etcd version to 3.4.4
- Kubelet now creates cgroup per QOS
- Kubelet now registers system reserved and kube reserved resources
- Kubelet now shares /run/xtables.lock with host to prevent races with kube-proxy
- All CLI tools now use generic code
- Re-enabled all linter warnings, which are disabled by default in golangci-lint and fixed found warnings
- update sonobuoy binary to 0.17.2 when running conformance tests in e2e environment
- Terraform code is now shared between local-testing and e2e environments when possible
- kube-apiserver will now validate kubelet's serving certificate
- Updated used Go version to 1.14

### Fixed
- When creating configuration files with Docker, they will have correct modification time now (#55)
- Trigger container updates when runtime configuration changes (#70)
- Removing containers will now properly remove all of them, not just first one (#75)
- Before doing actions on Helm releases, we will now make sure that API is reachable and ready. That fixes
flaky cluster deployments (#84)
- Etcd cluster now properly handles members with specified manual names
- containers won't be started, if they do not exist
- Docker runtime now properly finds if the image is pulled, even if image is not tagged
- containers will now be removed before they are upgraded to avoid conflicts
- containers which has updates pending will no longer be started, this allows to update containers with bad configuration
- Fixed reading status of config files to prevent unnecessary updates
- containers which are stopped won't be stopped before removing

### Removed
- .gitlab-ci.yml file, as it was added only experimentally and it was not used
- Image and Name fields from ContainerStatus, as they were not used

## [0.1.0] - 2020-01-28
### Added
- Initial release

[0.9.0]: https://github.com/flexkube/libflexkube/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/flexkube/libflexkube/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/flexkube/libflexkube/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/flexkube/libflexkube/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/flexkube/libflexkube/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/flexkube/libflexkube/compare/v0.4.3...v0.5.0
[0.4.3]: https://github.com/flexkube/libflexkube/compare/v0.4.2...v0.4.3
[0.4.2]: https://github.com/flexkube/libflexkube/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/flexkube/libflexkube/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/flexkube/libflexkube/compare/v0.3.3...v0.4.0
[0.3.3]: https://github.com/flexkube/libflexkube/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/flexkube/libflexkube/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/flexkube/libflexkube/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/flexkube/libflexkube/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/flexkube/libflexkube/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/flexkube/libflexkube/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/flexkube/libflexkube/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/flexkube/libflexkube/releases/tag/v0.1.0
