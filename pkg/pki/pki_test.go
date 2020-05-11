package pki

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	pki := &PKI{
		Etcd: &Etcd{
			Peers: map[string]string{
				"controller01": "192.168.1.10",
			},
			Servers: map[string]string{
				"controller01": "192.168.1.10",
			},
			ClientCNs: []string{
				"root",
				"kube-apiserver",
				"prometheus",
			},
		},
		Kubernetes: &Kubernetes{},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}
}

func TestGenerateNoConfig(t *testing.T) {
	t.Parallel()

	pki := &PKI{}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}
}

func TestGenerateBadRootCAPrivateKey(t *testing.T) {
	t.Parallel()

	pki := &PKI{
		RootCA: &Certificate{
			PrivateKey: "doh",
		},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail with invalid root private key")
	}
}

func TestGenerateBadEtcdCAPrivateKey(t *testing.T) {
	t.Parallel()

	pki := &PKI{
		Etcd: &Etcd{
			CA: &Certificate{
				PrivateKey: "doh",
			},
		},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail")
	}
}

func TestGenerateBadKubernetesCAPrivateKey(t *testing.T) {
	t.Parallel()

	pki := &PKI{
		Kubernetes: &Kubernetes{
			CA: &Certificate{
				PrivateKey: "doh",
			},
		},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail")
	}
}

func TestValidateValidityDuration(t *testing.T) {
	t.Parallel()

	pki := &PKI{
		RootCA: &Certificate{
			ValidityDuration: "doh",
		},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail")
	}
}

func TestValidateIPAddresses(t *testing.T) {
	t.Parallel()

	pki := &PKI{
		RootCA: &Certificate{
			IPAddresses: []string{"doh"},
		},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail")
	}
}

func TestDecodeX509CertificateNotPEM(t *testing.T) {
	t.Parallel()

	pki := &PKI{
		RootCA: &Certificate{
			X509Certificate: "foo",
		},
		Etcd: &Etcd{},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail on decoding Root CA certificate")
	}
}

func TestDecodeX509CertificateBadData(t *testing.T) {
	t.Parallel()

	pki := &PKI{
		RootCA: &Certificate{
			X509Certificate: `-----BEGIN CERTIFICATE-----
Zm9vCg==
-----END CERTIFICATE-----
`,
		},
		Etcd: &Etcd{},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail on decoding Root CA certificate")
	}
}

func TestGenerateEtcdCopyServers(t *testing.T) {
	t.Parallel()

	pki := &PKI{
		Etcd: &Etcd{
			Peers: map[string]string{
				"controller01": "192.168.1.10",
			},
		},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	if diff := cmp.Diff(pki.Etcd.Peers, pki.Etcd.Servers); diff != "" {
		t.Fatalf("servers should be copied from peers, if they are not defined, got: %v", diff)
	}
}
