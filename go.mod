module github.com/invidian/flexkube

go 1.13

require (
	cloud.google.com/go v0.47.0 // indirect
	cloud.google.com/go/storage v1.1.2 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/aws/aws-sdk-go v1.25.23 // indirect
	github.com/containerd/continuity v0.0.0-20190827140505-75bee3e2ccb6 // indirect
	github.com/docker/cli v0.0.0-20191030212255-1de18a9b9838 // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/gosuri/uitable v0.0.3 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-hclog v0.9.2 // indirect
	github.com/hashicorp/go-plugin v1.0.1 // indirect
	github.com/hashicorp/hcl2 v0.0.0-20191002203319-fb75b3253c80 // indirect
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.12.12
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/mattn/go-runewidth v0.0.5 // indirect
	github.com/miekg/dns v1.1.4 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/runc v1.0.0-rc2.0.20190611121236-6cc515888830 // indirect
	github.com/posener/complete v1.2.2 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/ulikunitz/xz v0.5.6 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/crypto v0.0.0-20191029031824-8986dd9e96cf
	golang.org/x/exp v0.0.0-20191030013958-a1ab85dbe136 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.0.0-20191029155521-f43be2a4598c // indirect
	golang.org/x/tools v0.0.0-20191030211004-889af361d29c // indirect
	google.golang.org/api v0.13.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20191026110619-0b21df46bc1d
	helm.sh/helm/v3 v3.0.0-rc.1
	k8s.io/apiextensions-apiserver v0.0.0-20191028232452-c47e10e6d5a3
	k8s.io/apimachinery v0.0.0-20191025225532-af6325b3a843
	k8s.io/client-go v0.0.0-20191028230319-1a481fb1e32d
	k8s.io/component-base v0.0.0-20191029070825-5e0e35147053 // indirect
	k8s.io/kubectl v0.0.0-20191027032433-6b77b0790ab4
	k8s.io/utils v0.0.0-20191030222137-2b95a09bc58d // indirect
	sigs.k8s.io/yaml v1.1.0
)

replace (
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2
	rsc.io/letsencrypt => github.com/dmcgowan/letsencrypt v0.0.0-20160928181947-1847a81d2087
)
