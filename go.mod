module github.com/flexkube/libflexkube

go 1.16

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/docker/docker v20.10.10+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/flexkube/helm/v3 v3.1.0-rc.1.0.20211028083037-3b856c17ab41
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/opencontainers/image-spec v1.0.2
	github.com/urfave/cli/v2 v2.3.0
	go.etcd.io/etcd/api/v3 v3.5.1
	go.etcd.io/etcd/client/v3 v3.5.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/component-base v0.23.0
	k8s.io/kube-scheduler v0.23.0
	k8s.io/kubectl v0.23.0
	k8s.io/kubelet v0.23.0
	sigs.k8s.io/yaml v1.3.0
)
