package etcd

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
)

func TestClusterFromYaml(t *testing.T) {
	c := `
ssh:
  user: "core"
  port: 2222
  password: foo
caCertificate: |
  {{.Certificate}}
members:
  foo:
    peerCertificate: |
      {{.MemberCertificate}}
    peerKey: |
      {{.MemberKey}}
    serverCertificate: |
      {{.MemberCertificate}}
    serverKey: |
      {{.MemberKey}}
    host:
      ssh:
        address: "127.0.0.1"
    peerAddress: 10.0.2.15
    serverAddress: 10.0.2.15
`

	data := struct {
		Certificate       string
		MemberCertificate string
		MemberKey         string
	}{
		strings.TrimSpace(util.Indent(utiltest.GenerateX509Certificate(t), "  ")),
		strings.TrimSpace(util.Indent(utiltest.GenerateX509Certificate(t), "      ")),
		strings.TrimSpace(util.Indent(utiltest.GenerateRSAPrivateKey(t), "      ")),
	}

	var buf bytes.Buffer

	tpl := template.Must(template.New("c").Parse(c))
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to generate config from template: %v", err)
	}

	if _, err := FromYaml(buf.Bytes()); err != nil {
		t.Fatalf("Creating etcd cluster from YAML should succeed, got: %v", err)
	}
}
