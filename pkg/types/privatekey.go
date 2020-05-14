package types

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strconv"

	"github.com/flexkube/libflexkube/internal/util"
)

// PrivateKey is a wrapper on string type, which parses it's content
// as private key while unmarshalling. This allows to validate the
// data during unmarshalling process.
type PrivateKey string

// UnmarshalJSON implements encoding/json.Unmarshaler interface.
func (p *PrivateKey) UnmarshalJSON(data []byte) error {
	up, err := strconv.Unquote(string(data))
	if err != nil {
		return fmt.Errorf("failed to unquote string: %v", err)
	}

	der, _ := pem.Decode([]byte(up))
	if der == nil {
		return fmt.Errorf("failed to decode PEM format")
	}

	if err := parsePrivateKey(der.Bytes); err != nil {
		return err
	}

	*p = PrivateKey(up)

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

	return fmt.Errorf("unable to parse private key")
}

// Pick returns first non-empty private key.
func (p *PrivateKey) Pick(values ...PrivateKey) PrivateKey {
	if p == nil || *p == "" {
		pt := PrivateKey("")
		p = &pt
	}

	pks := []string{string(*p)}
	for _, v := range values {
		pks = append(pks, string(v))
	}

	return PrivateKey(util.PickString(pks...))
}
