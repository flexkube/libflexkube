package pki_test

import (
	"testing"

	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/pki"
)

func TestPrivateKeyParse(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		key string
		err bool
	}{
		"bad": {
			"doh",
			true,
		},
		"badpem": {
			`
-----BEGIN CERTIFICATE-----
aHR0cHM6Ly93d3cueW91dHViZS5jb20vd2F0Y2g/dj1kUXc0dzlXZ1hjUQo=
-----END CERTIFICATE-----
`,
			true,
		},
		"pkcs1": {
			utiltest.GeneratePKCS1PrivateKey(t),
			false,
		},
		"ec": {
			utiltest.GenerateECPrivateKey(t),
			false,
		},
		"rsa": {
			utiltest.GenerateRSAPrivateKey(t),
			false,
		},
	}

	for n, testCase := range cases {
		testCase := testCase

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			err := pki.ValidatePrivateKey(testCase.key)
			if testCase.err && err == nil {
				t.Fatalf("Expected error and didn't get any.")
			}

			if !testCase.err && err != nil {
				t.Fatalf("Didn't expect error, got: %v", err)
			}
		})
	}
}
