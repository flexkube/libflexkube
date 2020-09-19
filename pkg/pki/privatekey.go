package pki

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// ValidatePrivateKey validates given private key in PEM format.
// If decoding or parsing fails, error is returned.
func ValidatePrivateKey(key string) error {
	der, _ := pem.Decode([]byte(key))
	if der == nil {
		return fmt.Errorf("decoding PEM format")
	}

	if err := parsePrivateKey(der.Bytes); err != nil {
		return fmt.Errorf("parsing private key: %w", err)
	}

	return nil
}

// parsePrivateKey tries to parse various private key types and
// returns error if none of them works.
func parsePrivateKey(b []byte) error {
	if _, err := x509.ParsePKCS8PrivateKey(b); err == nil {
		return nil
	}

	if _, err := x509.ParsePKCS1PrivateKey(b); err == nil {
		return nil
	}

	if _, err := x509.ParseECPrivateKey(b); err == nil {
		return nil
	}

	return fmt.Errorf("unsupported private key format, tried PKCS8, PKCS1 and EC formats")
}
