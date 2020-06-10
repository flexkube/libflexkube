package pki

import (
	"crypto/x509"
	"encoding/pem"
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

func TestGenerateTrustChain(t *testing.T) {
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

	roots := x509.NewCertPool()

	ok := roots.AppendCertsFromPEM([]byte(pki.RootCA.X509Certificate))
	if !ok {
		t.Fatal("failed to parse root certificate")
	}

	intermediates := x509.NewCertPool()

	ok = intermediates.AppendCertsFromPEM([]byte(pki.Etcd.CA.X509Certificate))
	if !ok {
		t.Fatal("failed to parse etcd CA certificate")
	}

	block, _ := pem.Decode([]byte(pki.Etcd.PeerCertificates["controller01"].X509Certificate))
	if block == nil {
		t.Fatal("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		DNSName:       "controller01",
		Intermediates: intermediates,
	}

	if _, err := cert.Verify(opts); err != nil {
		t.Fatalf("failed to verify certificate: %v", err)
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

func TestDecodeKeypair(t *testing.T) {
	ca := &Certificate{
		PrivateKey: "foo",
	}

	c := &Certificate{
		ValidityDuration: "24h",
		RSABits:          2048,
	}

	if err := c.Generate(ca); err == nil {
		t.Fatalf("generating certificate with bad CA should fail")
	}
}

func TestValidateRSABits(t *testing.T) {
	c := &Certificate{
		ValidityDuration: "24h",
	}

	if err := c.Validate(); err == nil {
		t.Fatalf("certificate with 0 RSA bits should be invalid")
	}
}
