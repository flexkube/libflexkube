package kubelet

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/pki"
	"github.com/flexkube/libflexkube/pkg/types"
)

func GetPool(t *testing.T) types.Resource {
	t.Helper()

	y := `
ssh:
  address: localhost
  password: foo
  connectionTimeout: 1s
  retryTimeout: 1s
  retryInterval: 1s
bootstrapConfig:
  server: "foo"
  token: "foo"
volumePluginDir: /var/lib/kubelet/volumeplugins
extraMounts:
- source: /foo/
  target: /bar
kubernetesCACertificate: |
  {{.}}
waitForNodeReady: false
extraArgs:
- --baz
kubelets:
- networkPlugin: cni
  name: foo
- networkPlugin: cni
  name: bar
  extraMounts:
  - source: /doh/
    target: /tmp
  extraArgs:
  - --bar
`

	var buf bytes.Buffer

	tpl := template.Must(template.New("c").Parse(y))
	if err := tpl.Execute(&buf, strings.TrimSpace(util.Indent(utiltest.GenerateX509Certificate(t), "  "))); err != nil {
		t.Fatalf("Failed to generate config from template: %v", err)
	}

	p, err := FromYaml(buf.Bytes())
	if err != nil {
		t.Fatalf("Creating pool from YAML should succeed, got: %v", err)
	}

	return p
}

// New() tests.
func TestPoolNewValidate(t *testing.T) {
	t.Parallel()

	y := `
ssh:
  address: localhost
  password: foo
  connectionTimeout: 1s
  retryTimeout: 1s
  retryInterval: 1s
volumePluginDir: /var/lib/kubelet/volumeplugins
kubelets:
- networkPlugin: cni
  name: foo
`

	if _, err := FromYaml([]byte(y)); err == nil {
		t.Fatalf("Creating pool from bad YAML should fail")
	}
}

// FromYaml() tests.
func TestPoolFromYaml(t *testing.T) {
	t.Parallel()

	GetPool(t)
}

// StateToYaml() tests.
func TestPoolStateToYAML(t *testing.T) {
	t.Parallel()

	p := GetPool(t)

	if _, err := p.StateToYaml(); err != nil {
		t.Fatalf("Dumping state to YAML should work, got: %v", err)
	}
}

// CheckCurrentState() tests.
func TestPoolCheckCurrentState(t *testing.T) {
	t.Parallel()

	p := GetPool(t)

	if err := p.CheckCurrentState(); err != nil {
		t.Fatalf("Checking current state of empty pool should work, got: %v", err)
	}
}

// Containers() tests.
func TestPoolContainers(t *testing.T) {
	t.Parallel()

	p := GetPool(t)

	if c := p.Containers(); c == nil {
		t.Fatalf("Containers() should return non-nil value")
	}
}

// Deploy() tests.
func TestPoolDeploy(t *testing.T) {
	t.Parallel()

	p := GetPool(t)

	if err := p.Deploy(); err == nil {
		t.Fatalf("Deploying in testing environment should fail")
	}
}

func TestPoolPropagateExtraMounts(t *testing.T) {
	t.Parallel()

	p := GetPool(t).(*pool)

	found := false

	for _, v := range p.containers.DesiredState()["0"].Container.Config.Mounts {
		if v.Source == "/foo/" && v.Target == "/bar" {
			found = true
		}
	}

	if !found {
		t.Errorf("kubelet foo should have propagated extra mount")
	}

	found = false

	for _, v := range p.containers.DesiredState()["1"].Container.Config.Mounts {
		if v.Source == "/doh/" && v.Target == "/tmp" {
			found = true
		}

		if v.Source == "/foo/" && v.Target == "/bar" {
			t.Errorf("kubelet doh should not have propagated mounts")
		}
	}

	if !found {
		t.Fatalf("kubelet doh should have directly configured extra mount")
	}
}

func Test_Pool_does_propagate_extra_args_when_instance_has_no_extra_args_set(t *testing.T) {
	t.Parallel()

	p := GetPool(t).(*pool)

	found := false

	for _, v := range p.containers.DesiredState()["0"].Container.Config.Args {
		if v == "--baz" {
			found = true
		}
	}

	if !found {
		t.Errorf("kubelet foo should have propagated extra arguments")
	}
}

func Test_Pool_does_preserve_extra_args_defined_in_instance(t *testing.T) {
	t.Parallel()

	p := GetPool(t).(*pool)

	found := false

	for _, arg := range p.containers.DesiredState()["1"].Container.Config.Args {
		if arg == "--bar" {
			found = true
		}

		if arg == "--baz" {
			t.Errorf("kubelet doh should not have propagated arguments")
		}
	}

	if !found {
		t.Fatalf("kubelet doh should have directly configured extra arguments")
	}
}

func TestPoolPKIIntegration(t *testing.T) {
	t.Parallel()

	pk := &pki.PKI{
		Kubernetes: &pki.Kubernetes{},
	}

	if err := pk.Generate(); err != nil {
		t.Fatalf("generating PKI: %v", err)
	}

	p := &Pool{
		PKI: pk,
		AdminConfig: &client.Config{
			Server: "foo",
		},
		BootstrapConfig: &client.Config{
			Server: "bar",
			Token:  "bar",
		},
		WaitForNodeReady: true,
		Kubelets: []Kubelet{
			{
				Name:            "foo",
				VolumePluginDir: "foo",
				NetworkPlugin:   "cni",
			},
		},
		PrivilegedLabels: map[string]string{
			"foo": "bar",
		},
	}

	if _, err := p.New(); err != nil {
		t.Fatalf("creating kubelet pool with PKI integration should work, got: %v", err)
	}
}

func TestPoolNoKubelets(t *testing.T) {
	t.Parallel()

	pk := &pki.PKI{
		Kubernetes: &pki.Kubernetes{},
	}

	if err := pk.Generate(); err != nil {
		t.Fatalf("generating PKI: %v", err)
	}

	p := &Pool{
		PKI: pk,
		BootstrapConfig: &client.Config{
			Server: "bar",
			Token:  "bar",
		},
	}

	if _, err := p.New(); err == nil {
		t.Fatal("creating kubelet pool with no kubelets and no state defined should fail")
	}
}
