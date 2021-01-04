// Package utiltest provides testing helpers, for generating valid mock data like
// X509 certificates etc.
package utiltest

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"
)

const (
	// certValidityDuration is how long certificate is valid from the moment of generation.
	certValidityDuration = 1 * time.Hour
)

// PKI struct holds X509 certificate and belonging RSA private key in PEM format.
type PKI struct {
	Certificate string
	PrivateKey  string
}

// GenerateX509Certificate generates random X.509 certificate and
// returns it as string in PEM format.
func GenerateX509Certificate(t *testing.T) string {
	t.Helper()

	return GeneratePKI(t).Certificate
}

// GenerateRSAPrivateKey generates RSA private key and returns it
// as string in PEM format.
func GenerateRSAPrivateKey(t *testing.T) string {
	t.Helper()

	return GeneratePKI(t).PrivateKey
}

// GeneratePKCS1PrivateKey generates RSA private key in PKCS1 format,
// PEM encoded.
func GeneratePKCS1PrivateKey(t *testing.T) string {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(priv)

	var key bytes.Buffer
	if err := pem.Encode(&key, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		t.Fatalf("Failed to write data to key.pem: %s", err)
	}

	return key.String()
}

// GenerateECPrivateKey generates EC private key, PEM encoded.
func GenerateECPrivateKey(t *testing.T) string {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed generating ECDSA key: %v", err)
	}

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("Failed serializing EC private key: %v", err)
	}

	var key bytes.Buffer
	if err := pem.Encode(&key, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}); err != nil {
		t.Fatalf("Failed to write data to key.pem: %s", err)
	}

	return key.String()
}

// GeneratePKI generates PKI struct.
func GeneratePKI(t *testing.T) *PKI {
	t.Helper()

	p, err := GeneratePKIErr()
	if err != nil {
		t.Fatalf("failed generating fake PKI: %v", err)
	}

	return p
}

// generateX509Certificate generates X.509 certificate in DER format using given RSA private key.
func generateX509Certificate(priv *rsa.PrivateKey) ([]byte, error) {
	// Generate serial number for X.509 certificate.
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"example"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(certValidityDuration),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create X.509 certificate in DER format.
	return x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
}

// encodePKI converts RSA private key and X.509 certificate in DER format into PKI struct.
func encodePKI(priv *rsa.PrivateKey, pub []byte) (*PKI, error) {
	// Encode private certificate into PEM format.
	var cert bytes.Buffer
	if err := pem.Encode(&cert, &pem.Block{Type: "CERTIFICATE", Bytes: pub}); err != nil {
		return nil, fmt.Errorf("failed to write data to cert.pem: %w", err)
	}

	// Convert RSA private key into PKCS8 DER format.
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal private key: %w", err)
	}

	// Convert private key from PKCS8 DER format to PEM format.
	var key bytes.Buffer
	if err := pem.Encode(&key, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return nil, fmt.Errorf("failed to write data to key.pem: %w", err)
	}

	return &PKI{
		Certificate: cert.String(),
		PrivateKey:  key.String(),
	}, nil
}

// GeneratePKIErr generates fake PKI X.509 key pair sutiable for tests.
func GeneratePKIErr() (*PKI, error) {
	// Generate RSA private key.
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	derBytes, err := generateX509Certificate(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create X.509 certificate: %w", err)
	}

	return encodePKI(priv, derBytes)
}
