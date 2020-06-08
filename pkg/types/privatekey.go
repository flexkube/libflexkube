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
//
// This type is deprecated, as it does not allow to produce
// meaningful error to the user.
type PrivateKey string

// UnmarshalJSON implements encoding/json.Unmarshaler interface and tries
// to decode obtained data using PEM format and then tries to parse the
// private key as PKCS8, PPKCS1 or EC private keys.
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

// Pick returns first non-empty private key from given list, including
// receiver private key.
//
// This method is a helper, which allows to select the certificate to use
// from hierarchical configuration.
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
