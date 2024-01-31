module github.com/flexkube/libflexkube

go 1.16

require (
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/docker/docker v23.0.8+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/flexkube/helm/v3 v3.1.0-rc.1.0.20230826150354-73f6b8d7f117
	github.com/google/go-cmp v0.5.9
	github.com/google/uuid v1.3.1
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/opencontainers/image-spec v1.1.0-rc4
	github.com/urfave/cli/v2 v2.25.7
	go.etcd.io/etcd/api/v3 v3.5.9
	go.etcd.io/etcd/client/v3 v3.5.9
	golang.org/x/crypto v0.17.0
	k8s.io/api v0.28.1
	k8s.io/apimachinery v0.28.1
	k8s.io/client-go v0.28.1
	k8s.io/component-base v0.28.1
	k8s.io/kube-scheduler v0.28.1
	k8s.io/kubectl v0.28.1
	k8s.io/kubelet v0.28.1
	sigs.k8s.io/yaml v1.3.0
)
