module github.com/flexkube/libflexkube

go 1.16

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/docker/docker v20.10.5+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/go-openapi/spec v0.20.3 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.2.0
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/opencontainers/image-spec v1.0.1
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/urfave/cli/v2 v2.3.0
	go.etcd.io/etcd/api/v3 v3.5.0
	go.etcd.io/etcd/client/v3 v3.5.0
	go.uber.org/atomic v1.8.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/sys v0.0.0-20210616045830-e2b7044e8c71 // indirect
	google.golang.org/genproto v0.0.0-20210614182748-5b3b54cad159 // indirect
	helm.sh/helm/v3 v3.6.0
	k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver v0.21.0
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
