module github.com/flexkube/libflexkube

go 1.16

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/docker/docker v20.10.17+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/flexkube/helm/v3 v3.1.0-rc.1.0.20220909120838-ecd8a196773e
	github.com/google/go-cmp v0.5.9
	github.com/google/uuid v1.3.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/opencontainers/image-spec v1.0.3-0.20211202183452-c5a74bcca799
	github.com/urfave/cli/v2 v2.14.1
	go.etcd.io/etcd/api/v3 v3.5.1
	go.etcd.io/etcd/client/v3 v3.5.1
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90
	k8s.io/api v0.25.0
	k8s.io/apimachinery v0.25.0
	k8s.io/client-go v0.25.0
	k8s.io/component-base v0.25.0
	k8s.io/kube-scheduler v0.25.0
	k8s.io/kubectl v0.25.0
	k8s.io/kubelet v0.25.0
	sigs.k8s.io/yaml v1.3.0
)
