module github.com/flexkube/libflexkube

go 1.16

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/docker/docker v20.10.5+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.2.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/opencontainers/image-spec v1.0.1
	github.com/urfave/cli/v2 v2.3.0
	go.etcd.io/etcd/api/v3 v3.5.0
	go.etcd.io/etcd/client/v3 v3.5.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	helm.sh/helm/v3 v3.6.3
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/kubectl v0.21.1
	k8s.io/kubelet v0.21.1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	// Borrowed from Helm.
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d

	// Use forked version of Helm to workaround https://github.com/helm/helm/issues/9761.
	helm.sh/helm/v3 => github.com/flexkube/helm/v3 v3.1.0-rc.1.0.20210728081922-539dfe1e558a
)
