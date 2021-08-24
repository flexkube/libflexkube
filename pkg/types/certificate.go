package types

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strconv"

	"github.com/flexkube/libflexkube/internal/util"
)

// Certificate is a wrapper on string type, which parses it's content
// as X.509 certificate while unmarshalling. This allows to validate the
// data during unmarshalling process.
//
// This type should not be used, as it does not allow to produce
// meaningful error to the user.
type Certificate string

// UnmarshalJSON implements encoding/json.Unmarshaler interface and tries
// to parse obtained data as PEM encoded X.509 certificate.
func (c *Certificate) UnmarshalJSON(data []byte) error {
	p, err := strconv.Unquote(string(data))
	if err != nil {
		return fmt.Errorf("unquoting string: %w", err)
	}

	der, _ := pem.Decode([]byte(p))
	if der == nil {
		return fmt.Errorf("decoding PEM format")
	}

	if _, err := x509.ParseCertificate(der.Bytes); err != nil {
		return fmt.Errorf("parsing certificate: %w", err)
	}

	*c = Certificate(p)

	return nil
}

// Pick returns first non-empty certificate from given list, including
// receiver certificate.
//
// This method is a helper, which allows to select the certificate to use
// from hierarchical configuration.
func (c *Certificate) Pick(values ...Certificate) Certificate {
	if c == nil || *c == "" {
		ce := Certificate("")
		c = &ce
	}

	cs := []string{string(*c)}
	for _, v := range values {
		cs = append(cs, string(v))
	}

	return Certificate(util.PickString(cs...))
}
