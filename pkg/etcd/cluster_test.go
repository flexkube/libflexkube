package etcd

import (
	"testing"
)

func TestClusterFromYaml(t *testing.T) {
	c := `
ssh:
  user: "core"
  port: 2222
  privateKey: foo
caCertificate: foo
members:
  foo:
    peerCertificate: foo
    peerKey: foo
    serverCertificate: foo
    serverKey: foo
    host:
      ssh:
        address: "127.0.0.1"
    peerAddress: 10.0.2.15
    serverAddress: 10.0.2.15
`

	if _, err := FromYaml([]byte(c)); err != nil {
		t.Fatalf("Creating etcd cluster from YAML should succeed, got: %v", err)
	}
}
