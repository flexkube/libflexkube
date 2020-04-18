package controlplane

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
)

func controlplaneYAML(t *testing.T) string {
	c := `
common:
  kubernetesCACertificate: |
    {{.Certificate}}
  frontProxyCACertificate: |
    {{.Certificate}}
kubeAPIServer:
  apiServerCertificate: |
    {{.Certificate}}
  apiServerKey: |
    {{.PrivateKey}}
  frontProxyCertificate: |
    {{.Certificate}}
  frontProxyKey: |
    {{.PrivateKey}}
  kubeletClientCertificate: |
    {{.Certificate}}
  kubeletClientKey: |
    {{.PrivateKey}}
  serviceAccountPublicKey: foo
  serviceCIDR: 11.0.0.0/24
  etcdCACertificate: |
    {{.Certificate}}
  etcdClientCertificate: |
    {{.Certificate}}
  etcdClientKey: |
    {{.PrivateKey}}
  etcdServers:
  - http://10.0.2.15:2379
  bindAddress: 0.0.0.0
  advertiseAddress: 127.0.0.1
kubeControllerManager:
  kubernetesCAKey: |
    {{.PrivateKey}}
  serviceAccountPrivateKey: |
    {{.PrivateKey}}
  kubeconfig:
    clientCertificate: |
      {{.CertificateDeep}}
    clientKey: |
      {{.PrivateKeyDeep}}
  rootCACertificate: |
    {{.Certificate}}
apiServerAddress: 127.0.0.1
kubeScheduler:
  kubeconfig:
    clientCertificate: |
      {{.CertificateDeep}}
    clientKey: |
      {{.PrivateKeyDeep}}
apiServerPort: 6443
ssh:
  user: "core"
  address: 127.0.0.1
  port: 2222
  password: "foo"
  connectionTimeout: 1ms
  retryTimeout: 1ms
  retryInterval: 1ms
`
	pki := utiltest.GeneratePKI(t)

	data := struct {
		Certificate     string
		PrivateKey      string
		CertificateDeep string
		PrivateKeyDeep  string
	}{
		strings.TrimSpace(util.Indent(utiltest.GenerateX509Certificate(t), "    ")),
		strings.TrimSpace(util.Indent(utiltest.GenerateRSAPrivateKey(t), "    ")),
		strings.TrimSpace(util.Indent(pki.Certificate, "      ")),
		strings.TrimSpace(util.Indent(pki.PrivateKey, "      ")),
	}

	var buf bytes.Buffer

	tpl := template.Must(template.New("c").Parse(c))
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to generate config from template: %v", err)
	}

	return buf.String()
}

func TestControlplaneFromYaml(t *testing.T) {
	co, err := FromYaml([]byte(controlplaneYAML(t)))
	if err != nil {
		t.Fatalf("Creating controlplane from YAML should succeed, got: %v", err)
	}

	if cc := co.Containers(); cc == nil {
		t.Fatalf("Containers() should return non-nil value")
	}

	if _, err := co.StateToYaml(); err != nil {
		t.Fatalf("Dumping state to YAML should work, got: %v", err)
	}

	if err := co.CheckCurrentState(); err != nil {
		t.Fatalf("Checking current state of empty controlplane should work, got: %v", err)
	}

	if err := co.Deploy(); err == nil {
		t.Fatalf("Deploying in testing environment should fail")
	}
}

// GetImage()
func TestCommonGetImage(t *testing.T) {
	c := Common{}
	if a := c.GetImage(); a == "" {
		t.Fatalf("GetImage() should always return at least default image")
	}
}

func TestCommonGetImageSpecified(t *testing.T) {
	i := "foo"
	c := Common{
		Image: i,
	}

	if a := c.GetImage(); a != i {
		t.Fatalf("GetImage() should return specified image, if it's defined")
	}
}

// New()
func TestControlplaneNewValidate(t *testing.T) {
	c := &Controlplane{}

	if _, err := c.New(); err == nil {
		t.Fatalf("New should validate controlplane configuration and fail on empty one")
	}
}

func TestControlplaneDestroyNoState(t *testing.T) {
	y := controlplaneYAML(t)

	y += `destroy: true`

	if _, err := FromYaml([]byte(y)); err == nil {
		t.Fatalf("creating controlplane config to destroy without state should fail")
	}
}

func TestControlplaneDestroyValidateState(t *testing.T) {
	y := controlplaneYAML(t)

	y += `destroy: true
state:
  foo: {}
`

	if _, err := FromYaml([]byte(y)); err == nil {
		t.Fatalf("creating controlplane config to destroy with invalid state should fail")
	}
}

func TestControlplaneDestroyValidState(t *testing.T) {
	y := controlplaneYAML(t)

	y += `destroy: true
state:
  foo:
    host:
      direct: {}
    container:
      runtime:
        docker:
          host: unix:///nonexistent
      config:
        name: foo
        image: busybox
      status:
        id: foo
        status: running
`

	if _, err := FromYaml([]byte(y)); err != nil {
		t.Fatalf("creating controlplane config to destroy with valid state should succeed, got: %v", err)
	}
}
