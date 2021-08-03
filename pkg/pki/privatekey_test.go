package pki_test

import (
	"testing"

	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/pki"
)

func Test_Validating_private_key(t *testing.T) {
	t.Parallel()

	t.Run("returns_error_when_given", func(t *testing.T) {
		cases := map[string]string{
			"non_PEM_encoded_input": "doh",
			"malformed_PEM_encoded_input": `
-----BEGIN CERTIFICATE-----
aHR0cHM6Ly93d3cueW91dHViZS5jb20vd2F0Y2g/dj1kUXc0dzlXZ1hjUQo=
-----END CERTIFICATE-----
`,
		}

		for n, key := range cases {
			key := key

			t.Run(n, func(t *testing.T) {
				t.Parallel()

				if err := pki.ValidatePrivateKey(key); err == nil {
					t.Fatalf("Expected error and didn't get any.")
				}
			})
		}
	})

	t.Run("accepts_PEM_encoded", func(t *testing.T) {
		t.Parallel()

		cases := map[string]string{
			"PKCS1_keys": utiltest.GeneratePKCS1PrivateKey(t),
			"EC_keys":    utiltest.GenerateECPrivateKey(t),
			"RSA_keys":   utiltest.GenerateRSAPrivateKey(t),
		}

		for n, key := range cases {
			key := key

			t.Run(n, func(t *testing.T) {
				t.Parallel()

				if err := pki.ValidatePrivateKey(key); err != nil {
					t.Fatalf("Unexpeced validation error: %v", err)
				}
			})
		}
	})
}
