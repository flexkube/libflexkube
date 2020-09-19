package pki_test

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/pki"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		Etcd: &pki.Etcd{
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
		Kubernetes: &pki.Kubernetes{},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}
}

func TestGenerateDontCopyAllSettings(t *testing.T) {
	t.Parallel()

	pkii := &pki.PKI{
		Kubernetes: &pki.Kubernetes{
			KubeAPIServer: &pki.KubeAPIServer{
				ServerIPs: []string{"1.1.1.1"},
			},
		},
	}

	if err := pkii.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	c := &pki.Certificate{
		X509Certificate: pkii.Kubernetes.KubeAPIServer.ServerCertificate.X509Certificate,
		PrivateKey:      pkii.Kubernetes.KubeAPIServer.ServerCertificate.PrivateKey,
		PublicKey:       pkii.Kubernetes.KubeAPIServer.ServerCertificate.PublicKey,
	}

	if diff := cmp.Diff(pkii.Kubernetes.KubeAPIServer.ServerCertificate, c); diff != "" {
		t.Fatalf("generated certificate should only have X.509 certificate and private key field populated, got: %v", diff)
	}
}

func TestGenerateTrustChain(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		Etcd: &pki.Etcd{
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
		// It seems staticcheck linter do not recognize t.Fatal() as a flow breaking statement,
		// so then it yells that we might dereference uninitialized 'block' variable, which is not
		// true, so just use t.Fatalf() to silence it.
		//
		// The alternative would be to add bare 'return' after this call, which seems even more ugly.
		t.Fatalf("failed to parse certificate PEM")
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

	pki := &pki.PKI{}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}
}

func TestGenerateBadRootCAPrivateKey(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		RootCA: &pki.Certificate{
			PrivateKey: "doh",
		},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail with invalid root private key")
	}
}

func TestGenerateBadEtcdCAPrivateKey(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		Etcd: &pki.Etcd{
			CA: &pki.Certificate{
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

	pki := &pki.PKI{
		Kubernetes: &pki.Kubernetes{
			CA: &pki.Certificate{
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

	pki := &pki.PKI{
		RootCA: &pki.Certificate{
			ValidityDuration: "doh",
		},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail")
	}
}

func TestValidateIPAddresses(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		RootCA: &pki.Certificate{
			IPAddresses: []string{"doh"},
		},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail")
	}
}

func TestDecodeX509CertificateNotPEM(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		RootCA: &pki.Certificate{
			X509Certificate: "foo",
		},
		Etcd: &pki.Etcd{},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail on decoding Root CA certificate")
	}
}

func TestDecodeX509CertificateBadData(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		RootCA: &pki.Certificate{
			X509Certificate: `-----BEGIN CERTIFICATE-----
Zm9vCg==
-----END CERTIFICATE-----
`,
		},
		Etcd: &pki.Etcd{},
	}

	if err := pki.Generate(); err == nil {
		t.Fatalf("generating should fail on decoding Root CA certificate")
	}
}

func TestGenerateEtcdCopyServers(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		Etcd: &pki.Etcd{
			Peers: map[string]string{
				"controller01": "192.168.1.10",
			},
		},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	if len(pki.Etcd.ServerCertificates) == 0 {
		t.Fatalf("if servers are not defined, certificates should be created from peers")
	}
}

func TestDecodeKeypair(t *testing.T) {
	ca := &pki.Certificate{
		PrivateKey: "foo",
	}

	c := &pki.Certificate{
		ValidityDuration: "24h",
		RSABits:          2048,
	}

	if err := c.Generate(ca); err == nil {
		t.Fatalf("generating certificate with bad CA should fail")
	}
}

func TestValidateRSABits(t *testing.T) {
	c := &pki.Certificate{
		ValidityDuration: "24h",
	}

	if err := c.Validate(); err == nil {
		t.Fatalf("certificate with 0 RSA bits should be invalid")
	}
}

func TestGenerateUpdateIPs(t *testing.T) {
	t.Parallel()

	// First, generate valid PKI.
	pki := &pki.PKI{
		Kubernetes: &pki.Kubernetes{
			KubeAPIServer: &pki.KubeAPIServer{
				ServerIPs: []string{"1.1.1.1"},
			},
		},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	// Save content of generated certificate.
	cert := pki.Kubernetes.KubeAPIServer.ServerCertificate.X509Certificate

	// Update ServerIPs.
	pki.Kubernetes.KubeAPIServer.ServerIPs = []string{"1.1.1.1", "2.2.2.2"}

	// Generate again to update the certificate.
	if err := pki.Generate(); err != nil {
		t.Fatalf("re-generating PKI certificates should succeed, got: %v", err)
	}

	if cert == pki.Kubernetes.KubeAPIServer.ServerCertificate.X509Certificate {
		t.Fatalf("certificate should be updated when IP addresses change")
	}
}

func TestGenerateDontRecreate(t *testing.T) {
	t.Parallel()

	// First, generate valid PKI.
	pki := &pki.PKI{}

	if err := pki.Generate(); err != nil {
		t.Fatalf("generating valid PKI should work, got: %v", err)
	}

	// Save content of generated certificate.
	cert := pki.RootCA.X509Certificate

	// Generate again.
	if err := pki.Generate(); err != nil {
		t.Fatalf("re-generating PKI certificates should succeed, got: %v", err)
	}

	if cert != pki.RootCA.X509Certificate {
		t.Fatalf("with no configuration changes, certificates should not be rotated")
	}
}

func TestIsX509CertificateUpToDateBadCert(t *testing.T) {
	t.Parallel()

	c := &pki.Certificate{
		X509Certificate: "doh",
	}

	if _, err := c.IsX509CertificateUpToDate(); err == nil {
		t.Fatalf("checking if certificate is up to date should fail on bad certificate")
	}
}
