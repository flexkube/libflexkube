package utiltest

import (
	"testing"
)

// GenerateX509Certificate()
func TestGenerateX509Certificate(t *testing.T) {
	if a := GenerateX509Certificate(t); a == "" {
		t.Fatalf("Generating X509 certificate should not return empty string")
	}
}

// GenerateRSAPrivateKey()
func TestGenerateRSAPrivateKey(t *testing.T) {
	if a := GenerateRSAPrivateKey(t); a == "" {
		t.Fatalf("Generating RSA private key should not return empty string")
	}
}

// GeneratePKI()
func TestGeneratePKI(t *testing.T) {
	p := GeneratePKI(t)

	if p.Certificate == "" {
		t.Errorf("PKI shouldn't have empty certificate field")
	}

	if p.PrivateKey == "" {
		t.Errorf("PKI shouldn't have empty private key")
	}
}
