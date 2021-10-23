module github.com/flexkube/libflexkube

go 1.16

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/docker/docker v20.10.8+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/flexkube/helm/v3 v3.1.0-rc.1.0.20210917141739-b8ae023d5dda
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/opencontainers/image-spec v1.0.1
	github.com/urfave/cli/v2 v2.3.0
	go.etcd.io/etcd/api/v3 v3.5.1
	go.etcd.io/etcd/client/v3 v3.5.1
	golang.org/x/crypto v0.0.0-20210915214749-c084706c2272
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/kubectl v0.22.1
	k8s.io/kubelet v0.22.2
	sigs.k8s.io/yaml v1.2.0
)
