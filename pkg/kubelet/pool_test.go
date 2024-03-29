package kubelet_test

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/kubelet"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/pki"
	"github.com/flexkube/libflexkube/pkg/types"
)

func getPool(t *testing.T) types.Resource {
	t.Helper()

	configTemplate := `
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
- name: foo
- name: bar
  extraMounts:
  - source: /doh/
    target: /tmp
  extraArgs:
  - --bar
`

	var buf bytes.Buffer

	tpl := template.Must(template.New("c").Parse(configTemplate))
	if err := tpl.Execute(&buf, strings.TrimSpace(util.Indent(utiltest.GenerateX509Certificate(t), "  "))); err != nil {
		t.Fatalf("Failed to generate config from template: %v", err)
	}

	p, err := kubelet.FromYaml(buf.Bytes())
	if err != nil {
		t.Fatalf("Creating pool from YAML should succeed, got: %v", err)
	}

	return p
}

// New() tests.
func TestPoolNewValidate(t *testing.T) {
	t.Parallel()

	testConfigRaw := `
ssh:
  address: localhost
  password: foo
  connectionTimeout: 1s
  retryTimeout: 1s
  retryInterval: 1s
volumePluginDir: /var/lib/kubelet/volumeplugins
kubelets:
- name: foo
`

	if _, err := kubelet.FromYaml([]byte(testConfigRaw)); err == nil {
		t.Fatalf("Creating pool from bad YAML should fail")
	}
}

// FromYaml() tests.
func TestPoolFromYaml(t *testing.T) {
	t.Parallel()

	getPool(t)
}

// StateToYaml() tests.
func TestPoolStateToYAML(t *testing.T) {
	t.Parallel()

	p := getPool(t)

	if _, err := p.StateToYaml(); err != nil {
		t.Fatalf("Dumping state to YAML should work, got: %v", err)
	}
}

// CheckCurrentState() tests.
func TestPoolCheckCurrentState(t *testing.T) {
	t.Parallel()

	p := getPool(t)

	if err := p.CheckCurrentState(); err != nil {
		t.Fatalf("Checking current state of empty pool should work, got: %v", err)
	}
}

// Containers() tests.
func TestPoolContainers(t *testing.T) {
	t.Parallel()

	p := getPool(t)

	if c := p.Containers(); c == nil {
		t.Fatalf("Containers() should return non-nil value")
	}
}

// Deploy() tests.
func TestPoolDeploy(t *testing.T) {
	t.Parallel()

	p := getPool(t)

	if err := p.Deploy(); err == nil {
		t.Fatalf("Deploying in testing environment should fail")
	}
}

func Test_Pool_propagates_extra_mounts_to_members_without_extra_mounts_defined(t *testing.T) {
	t.Parallel()

	p := getPool(t)

	found := false

	for _, v := range p.Containers().DesiredState()["0"].Container.Config.Mounts {
		if v.Source == "/foo/" && v.Target == "/bar" {
			found = true
		}
	}

	if !found {
		t.Fatal("Kubelet foo should have propagated extra mount")
	}
}

func Test_Pool_retains_individual_members_extra_mounts(t *testing.T) {
	t.Parallel()

	p := getPool(t)

	found := false

	for _, v := range p.Containers().DesiredState()["1"].Container.Config.Mounts {
		if v.Source == "/doh/" && v.Target == "/tmp" {
			found = true
		}

		if v.Source == "/foo/" && v.Target == "/bar" {
			t.Errorf("Kubelet doh should not have propagated mounts")
		}
	}

	if !found {
		t.Fatalf("Kubelet doh should have directly configured extra mount")
	}
}

func Test_Pool_does_propagate_extra_args_when_instance_has_no_extra_args_set(t *testing.T) {
	t.Parallel()

	p := getPool(t)

	found := false

	for _, v := range p.Containers().DesiredState()["0"].Container.Config.Args {
		if v == "--baz" {
			found = true
		}
	}

	if !found {
		t.Errorf("Kubelet foo should have propagated extra arguments")
	}
}

func Test_Pool_does_preserve_extra_args_defined_in_instance(t *testing.T) {
	t.Parallel()

	p := getPool(t)

	found := false

	for _, arg := range p.Containers().DesiredState()["1"].Container.Config.Args {
		if arg == "--bar" {
			found = true
		}

		if arg == "--baz" {
			t.Errorf("Kubelet doh should not have propagated arguments")
		}
	}

	if !found {
		t.Fatalf("Kubelet doh should have directly configured extra arguments")
	}
}

func TestPoolPKIIntegration(t *testing.T) {
	t.Parallel()

	testPKI := &pki.PKI{
		Kubernetes: &pki.Kubernetes{},
	}

	if err := testPKI.Generate(); err != nil {
		t.Fatalf("Generating PKI: %v", err)
	}

	pool := &kubelet.Pool{
		PKI: testPKI,
		AdminConfig: &client.Config{
			Server: "foo",
		},
		BootstrapConfig: &client.Config{
			Server: "bar",
			Token:  "bar",
		},
		WaitForNodeReady: true,
		Kubelets: []kubelet.Kubelet{
			{
				Name:            "foo",
				VolumePluginDir: "foo",
			},
		},
		PrivilegedLabels: map[string]string{
			"foo": "bar",
		},
	}

	if _, err := pool.New(); err != nil {
		t.Fatalf("Creating kubelet pool with PKI integration should work, got: %v", err)
	}
}

func TestPoolNoKubelets(t *testing.T) {
	t.Parallel()

	testPKI := &pki.PKI{
		Kubernetes: &pki.Kubernetes{},
	}

	if err := testPKI.Generate(); err != nil {
		t.Fatalf("Generating PKI: %v", err)
	}

	pool := &kubelet.Pool{
		PKI: testPKI,
		BootstrapConfig: &client.Config{
			Server: "bar",
			Token:  "bar",
		},
	}

	if _, err := pool.New(); err == nil {
		t.Fatal("Creating kubelet pool with no kubelets and no state defined should fail")
	}
}
