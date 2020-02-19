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
	return GeneratePKI(t).Certificate
}

// GenerateRSAPrivateKey generates RSA private key and returns it
// as string in PEM format.
func GenerateRSAPrivateKey(t *testing.T) string {
	return GeneratePKI(t).PrivateKey
}

// GeneratePKCS1PrivateKey generates RSA private key in PKCS1 format,
// PEM encoded.
func GeneratePKCS1PrivateKey(t *testing.T) string {
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

// GeneratePKI generates PKI struct
func GeneratePKI(t *testing.T) *PKI {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128) //nolint:gomnd

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		t.Fatalf("Failed to generate serial number: %v", err)
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

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	var cert bytes.Buffer
	if err := pem.Encode(&cert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		t.Fatalf("Failed to write data to cert.pem: %s", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("Unable to marshal private key: %v", err)
	}

	var key bytes.Buffer
	if err := pem.Encode(&key, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		t.Fatalf("Failed to write data to key.pem: %s", err)
	}

	return &PKI{
		Certificate: cert.String(),
		PrivateKey:  key.String(),
	}
}
