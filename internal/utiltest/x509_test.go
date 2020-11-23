package utiltest

import (
	"testing"
)

// GenerateX509Certificate() tests.
func TestGenerateX509Certificate(t *testing.T) {
	t.Parallel()

	if a := GenerateX509Certificate(t); a == "" {
		t.Fatalf("Generating X509 certificate should not return empty string")
	}
}

// GenerateRSAPrivateKey() tests.
func TestGenerateRSAPrivateKey(t *testing.T) {
	t.Parallel()

	if a := GenerateRSAPrivateKey(t); a == "" {
		t.Fatalf("Generating RSA private key should not return empty string")
	}
}

// GeneratePKI() tests.
func TestGeneratePKI(t *testing.T) {
	t.Parallel()

	p := GeneratePKI(t)

	if p.Certificate == "" {
		t.Errorf("PKI shouldn't have empty certificate field")
	}

	if p.PrivateKey == "" {
		t.Errorf("PKI shouldn't have empty private key")
	}
}
