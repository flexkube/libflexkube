package pki_test

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/pki"
)

func TestGenerateEtcdPeerCertificates(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		Etcd: &pki.Etcd{
			PeerCertificates: map[string]*pki.Certificate{
				"foo": {
					Organization: "foo",
				},
			},
		},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	if pki.Etcd.PeerCertificates["foo"].X509Certificate == "" {
		t.Fatalf("generated etcd peer certificate should not be empty")
	}
}

func TestGenerateEtcdPeerCertificatesPropagate(t *testing.T) {
	t.Parallel()

	e := []net.IP{net.ParseIP("1.1.1.1")}

	pki := &pki.PKI{
		Etcd: &pki.Etcd{
			Peers: map[string]string{
				"foo": "1.1.1.1",
			},
			PeerCertificates: map[string]*pki.Certificate{
				"foo": {
					Organization: "foo",
				},
			},
		},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	if pki.Etcd.PeerCertificates["foo"].X509Certificate == "" {
		t.Fatalf("generated etcd peer certificate should not be empty")
	}

	c, err := pki.Etcd.PeerCertificates["foo"].DecodeX509Certificate()
	if err != nil {
		t.Fatalf("decoding generated certificate should work, got: %v", err)
	}

	if diff := cmp.Diff(c.IPAddresses, e); diff != "" {
		t.Fatalf("unexpected diff: %s", diff)
	}
}

func TestGenerateEtcdPeerCertitificatesSupportAddingPeers(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		Etcd: &pki.Etcd{
			Peers: map[string]string{
				"foo": "1.1.1.1",
			},
			PeerCertificates: map[string]*pki.Certificate{
				"foo": {
					Organization: "foo",
				},
			},
		},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	pki.Etcd.Peers["bar"] = "2.2.2.2"

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	if pki.Etcd.PeerCertificates["bar"].X509Certificate == "" {
		t.Fatalf("generated etcd peer certificate should not be empty")
	}
}

func TestGenerateEtcdPeerCertitificatesPreservePeers(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		Etcd: &pki.Etcd{
			Peers: map[string]string{
				"foo": "1.1.1.1",
			},
		},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	pki.Etcd.Peers = map[string]string{}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	if pki.Etcd.PeerCertificates["foo"].X509Certificate == "" {
		t.Fatalf("generated etcd peer certificate should not be empty")
	}
}

func TestGenerateEtcdPeerCertitificatesAddServer(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		Etcd: &pki.Etcd{
			Peers: map[string]string{
				"foo": "1.1.1.1",
			},
		},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	pki.Etcd.Servers = map[string]string{"bar": "2.2.2.2"}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	if pki.Etcd.ServerCertificates["bar"].X509Certificate == "" {
		t.Fatalf("generated etcd server certificate should not be empty")
	}
}
