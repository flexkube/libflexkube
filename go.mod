module github.com/flexkube/libflexkube

go 1.16

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/docker/docker v20.10.5+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.2.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/urfave/cli/v2 v2.3.0
	go.etcd.io/etcd/api/v3 v3.5.0-alpha.0
	go.etcd.io/etcd/client/v3 v3.5.0-alpha.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	helm.sh/helm/v3 v3.5.3
	k8s.io/api v0.20.5
	k8s.io/apiextensions-apiserver v0.20.5
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v0.20.5
	k8s.io/kubectl v0.20.5
	k8s.io/kubelet v0.20.5
	sigs.k8s.io/yaml v1.2.0
)

replace (
	// Borrowed from Helm.
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d

	// Force updating docker/docker to v19.03.15.
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible

	// With v0.2.0 package has been renames, so until all dependencies are updated to use new import name,
	// we need to use older version.
	//
	// See: https://github.com/moby/spdystream/releases/tag/v0.2.0
	github.com/docker/spdystream => github.com/moby/spdystream v0.1.0

	// sigs.k8s.io/kustomize@v2.0.3+incompatible pulled by
	// k8s.io/cli-runtime pulled by helm.sh/helm/v3
	// is not compatible with spec v0.19.9.
	github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.8

	// k8s.io/kubectl is not compatible with never version.
	github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2
)
