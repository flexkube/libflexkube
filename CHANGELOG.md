# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.2] - 2020-04-19

### Added

- It is now possible to configure extra mounts for kubelet container via extraMounts/extra_mount parameters
- etcd is now ready for enabling RBAC
- local-testing environment now generates script for enabling etcd RBAC

### Fixed

- HAProxy now use HTTPS for probing kube-apiserver to avoid extensive logging of TLS handshake errors
- HAProxy configuration no longer generates warnings
- Fixed destroying flexkube_controlplane resource
- It is now possible to add and remove nodes in local-testing environment

### Changed

- Default Kubernetes version is now 1.18.2
- Default HAProxy version is now 2.1.4
- Default etcd version is now 3.4.7
- Improved validation rules of controlplane. Now state from previous deployments will be validated as well.


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
- Added support for specifing user and group when running containers (#57)
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
- Splitted README.md into smaller documents and added ToC for all of them (#87)
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

### Removed

- .gitlab-ci.yml file, as it was added only experimentally and it was not used
- Image and Name fields from ContainerStatus, as they were not used

### Fixed

- When creating configuration files with Docker, they will have correct modification time now (#55)
- Trigger container updates when runtime configuration changes (#70)
- Removing containes will now properly remove all of them, not just first one (#75)
- Before doing actions on Helm releases, we will now make sure that API is reachable and ready. That fixes
  flaky cluster deployments (#84)
- Etcd cluster now properly handles members with specified manual names
- containers won't be started, if they do not exist
- Docker runtime now properly finds if the image is pulled, even if image is not tagged
- containers will now be removed before they are upgraded to avoid conflicts
- containers which has updates pending will no longer be started, this allows to update containers with bad configuration
- Fixed reading status of config files to prevent unnecessary updates
- containers which are stopped won't be stopped before removing


## [0.1.0] - 2020-01-28

### Added

- Initial release

[0.2.0]: https://github.com/flexkube/libflexkube/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/flexkube/libflexkube/releases/tag/v0.1.0
