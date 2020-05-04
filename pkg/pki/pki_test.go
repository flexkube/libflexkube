package pki

import (
	"testing"
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
