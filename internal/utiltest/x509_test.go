package utiltest

import (
	"crypto/x509"
	"encoding/pem"
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

// GeneratePKCS1PrivateKey tests.
func Test_GeneratePKCS1PrivateKey_returns_PEM_encoded_RSA_private_key(t *testing.T) {
	t.Parallel()

	pemEncodedPrivateKey := GeneratePKCS1PrivateKey(t)

	derPrivateKey, _ := pem.Decode([]byte(pemEncodedPrivateKey))
	if derPrivateKey == nil {
		t.Fatalf("Returned key is not PEM encoded:\n%s", pemEncodedPrivateKey)
	}

	if _, err := x509.ParsePKCS1PrivateKey(derPrivateKey.Bytes); err != nil {
		t.Fatalf("Returned key is not PKCS1 private key")
	}
}

// GenerateECPrivateKey tests.
func Test_GenerateECPrivateKey_returns_PEM_encoded_EC_private_key(t *testing.T) {
	t.Parallel()

	pemEncodedPrivateKey := GenerateECPrivateKey(t)

	derPrivateKey, _ := pem.Decode([]byte(pemEncodedPrivateKey))
	if derPrivateKey == nil {
		t.Fatalf("Returned key is not PEM encoded:\n%s", pemEncodedPrivateKey)
	}

	if _, err := x509.ParseECPrivateKey(derPrivateKey.Bytes); err != nil {
		t.Fatalf("Returned key is not EC private key")
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
