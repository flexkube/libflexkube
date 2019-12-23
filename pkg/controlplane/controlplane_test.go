package controlplane

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
)

func TestControlplaneFromYaml(t *testing.T) {
	c := `kubernetesCACertificate: |
  {{.CertificateTop}}
frontProxyCACertificate: |
  {{.CertificateTop}}
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
  clientCertificate: |
    {{.Certificate}}
  clientKey: |
    {{.PrivateKey}}
  rootCACertificate: |
    {{.Certificate}}
  apiServer: 127.0.0.1:6443
kubeScheduler:
  clientCertificate: |
    {{.Certificate}}
  clientKey: |
    {{.PrivateKey}}
  apiServer: 127.0.0.1:6443
apiServerPort: 6443
ssh:
  user: "core"
  address: 127.0.0.1
  port: 2222
  password: "foo"
`
	data := struct {
		CertificateTop string
		Certificate    string
		PrivateKey     string
	}{
		strings.TrimSpace(util.Indent(utiltest.GenerateX509Certificate(t), "  ")),
		strings.TrimSpace(util.Indent(utiltest.GenerateX509Certificate(t), "    ")),
		strings.TrimSpace(util.Indent(utiltest.GenerateRSAPrivateKey(t), "    ")),
	}

	var buf bytes.Buffer

	tpl := template.Must(template.New("c").Parse(c))
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to generate config from template: %v", err)
	}

	if _, err := FromYaml(buf.Bytes()); err != nil {
		t.Fatalf("Creating controlplane from YAML should succeed, got: %v", err)
	}
}
