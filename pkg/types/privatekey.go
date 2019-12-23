package types

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strconv"
)

// PrivateKey is a wrapper on string type, which parses it's content
// as PKCS#1 private key while unmarshalling. This allows to validate the
// data during unmarshalling process.
type PrivateKey string

// UnmarshalJSON implements encoding/json.Unmarshaler interface.
func (c *PrivateKey) UnmarshalJSON(data []byte) error {
	p, err := strconv.Unquote(string(data))
	if err != nil {
		return fmt.Errorf("failed to unquote string: %v", err)
	}

	der, _ := pem.Decode([]byte(p))
	if der == nil {
		return fmt.Errorf("failed to decode PEM format")
	}

	if _, err := x509.ParsePKCS8PrivateKey(der.Bytes); err == nil {
		return nil
	}

	if _, err := x509.ParsePKCS1PrivateKey(der.Bytes); err == nil {
		return nil
	}

	if _, err := x509.ParseECPrivateKey(der.Bytes); err == nil {
		return nil
	}

	return fmt.Errorf("unable to parse private key")
}
