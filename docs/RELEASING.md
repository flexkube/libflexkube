# libflexkube release process

This document briefly describes the process of releasing new version of libflexkube.

## Before the release

Before creating a Git tag and a GitHub release, following tasks should be performed:

- Changelog for the new version should be added to CHANGELOG.md file.
- Changelog link should be added at the bottom of CHANGELOG.md file.
- `Version` constant in `cli/flexkube/cli.go` file should be changed to version which will be released and change should be committed. This commit will be later on tagged while releasing, so it should be done as last action before the release.
- `Version` constant in `cli/flexkube/cli.go` file should be changed to next version with `-unreleased` suffix and change should be committed. This commit will be the first commit of the next release.
- Conformance tests (e.g. `make vagrant-conformance`) should be performed on the release commit before creating an actual release to ensure the release is working properly.
- Before creating a Pull Request, run `goreleaser --skip-publish` to ensure that the release will build for all desired platforms.
- Pull Request with described changes should be created and merged.

## Creating the release

To create new release, following tasks should be performed:

- Tag new release on desired commit with CLI version changed, using example command:

  ```sh
  git tag -a v0.4.7 -s -m "Release v0.4.7" <commit hash>
  ```

- Push tag to GitHub:

  ```sh
  git push upstream v0.4.7
  ```

- Run `goreleaser` to create a GitHub Release:

  ```sh
  GITHUB_TOKEN=githubtoken goreleaser release --release-notes <(go run github.com/rcmachado/changelog show 0.4.7)
  ```

- Go to newly create [GitHub release](https://github.com/flexkube/libflexkube/releases/tag/v0.4.7), verify that the changelog and artifacts looks correct and publish it.
