module github.com/flexkube/libflexkube

go 1.13

require (
	cloud.google.com/go v0.49.0 // indirect
	cloud.google.com/go/storage v1.4.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/semver/v3 v3.0.2 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/aws/aws-sdk-go v1.25.44 // indirect
	github.com/bmatcuk/doublestar v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/containerd/containerd v1.3.1 // indirect
	github.com/containerd/continuity v0.0.0-20191127005431-f65d91d395eb // indirect
	github.com/deislabs/oras v0.8.0 // indirect
	github.com/docker/cli v0.0.0-20191126203649-54d085b857e9 // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-hclog v0.10.0 // indirect
	github.com/hashicorp/go-plugin v1.0.1 // indirect
	github.com/hashicorp/hcl/v2 v2.1.0 // indirect
	github.com/hashicorp/hcl2 v0.0.0-20191002203319-fb75b3253c80 // indirect
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.12.16
	github.com/hashicorp/terraform-svchost v0.0.0-20191119180714-d2e4933b9136 // indirect
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/huandu/xstrings v1.2.1 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mattn/go-runewidth v0.0.6 // indirect
	github.com/miekg/dns v1.1.4 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/runc v1.0.0-rc2.0.20190611121236-6cc515888830 // indirect
	github.com/posener/complete v1.2.3 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/ulikunitz/xz v0.5.6 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.1.1 // indirect
	go.opencensus.io v0.22.2 // indirect
	golang.org/x/crypto v0.0.0-20191202143827-86a70503ff7e
	golang.org/x/exp v0.0.0-20191129062945-2f5052295587 // indirect
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/net v0.0.0-20191126235420-ef20fe5d7933 // indirect
	golang.org/x/oauth2 v0.0.0-20191122200657-5d9234df094c // indirect
	golang.org/x/sys v0.0.0-20191128015809-6d18c012aee9 // indirect
	golang.org/x/tools v0.0.0-20191130070609-6e064ea0cf2d // indirect
	google.golang.org/genproto v0.0.0-20191115221424-83cc0476cb11 // indirect
	google.golang.org/grpc v1.25.1 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
	helm.sh/helm/v3 v3.0.0
	k8s.io/apiextensions-apiserver v0.0.0-20191121021419-88daf26ec3b8
	k8s.io/apimachinery v0.0.0-20191123233150-4c4803ed55e3
	k8s.io/client-go v0.0.0-20191123055820-8d0e6f1b7b78
	k8s.io/component-base v0.0.0-20191126060427-9e57f2cf1eda // indirect
	k8s.io/kubectl v0.0.0-20191126022606-c273777957bd
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
	sigs.k8s.io/yaml v1.1.0
)

replace (
	github.com/deislabs/oras => github.com/deislabs/oras v0.7.0
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2
	rsc.io/letsencrypt => github.com/dmcgowan/letsencrypt v0.0.0-20160928181947-1847a81d2087
)
