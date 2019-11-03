module github.com/invidian/libflexkube

go 1.13

require (
	cloud.google.com/go v0.47.0 // indirect
	cloud.google.com/go/storage v1.1.2 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/aws/aws-sdk-go v1.25.25 // indirect
	github.com/containerd/continuity v0.0.0-20190827140505-75bee3e2ccb6 // indirect
	github.com/deislabs/oras v0.8.0 // indirect
	github.com/docker/cli v0.0.0-20191031185610-968ce1ae4d45 // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/gosuri/uitable v0.0.3 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-hclog v0.10.0 // indirect
	github.com/hashicorp/go-plugin v1.0.1 // indirect
	github.com/hashicorp/hcl2 v0.0.0-20191002203319-fb75b3253c80 // indirect
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.12.13
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
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
	golang.org/x/net v0.0.0-20191101175033-0deb6923b6d9 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.0.0-20191029155521-f43be2a4598c // indirect
	golang.org/x/tools v0.0.0-20191101200257-8dbcdeb83d3f // indirect
	google.golang.org/api v0.13.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20191026110619-0b21df46bc1d
	helm.sh/helm/v3 v3.0.0-rc.2
	k8s.io/api v0.0.0-20191102065807-b98ecd433b91 // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20191101231532-ec619caef027
	k8s.io/apimachinery v0.0.0-20191102025618-50aa20a7b23f
	k8s.io/cli-runtime v0.0.0-20191102031428-d1199d98239f // indirect
	k8s.io/client-go v0.0.0-20191101230044-e9766ae82012
	k8s.io/kubectl v0.0.0-20191101233251-e953b35e53e0
	sigs.k8s.io/yaml v1.1.0
)

replace (
	github.com/deislabs/oras => github.com/deislabs/oras v0.7.0
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2
	rsc.io/letsencrypt => github.com/dmcgowan/letsencrypt v0.0.0-20160928181947-1847a81d2087
)
