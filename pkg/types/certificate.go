package types

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strconv"
)

// Certificate is a wrapper on string type, which parses it's content
// as X.509 certificate while unmarshalling. This allows to validate the
// data during unmarshalling process.
type Certificate string

// UnmarshalJSON implements encoding/json.Unmarshaler interface.
func (c *Certificate) UnmarshalJSON(data []byte) error {
	p, err := strconv.Unquote(string(data))
	if err != nil {
		return fmt.Errorf("failed to unquote string: %v", err)
	}

	der, _ := pem.Decode([]byte(p))
	if der == nil {
		return fmt.Errorf("failed to decode PEM format")
	}

	if _, err := x509.ParseCertificate(der.Bytes); err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	*c = Certificate(p)

	return nil
}
