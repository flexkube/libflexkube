package types_test

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/types"
)

type Foo struct {
	Bar types.PrivateKey `json:"bar"`
}

func TestPrivateKeyParse(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		YAML  string
		Error bool
	}{
		"bad": {
			"bar: doh",
			true,
		},
		"good": {
			fmt.Sprintf("bar: |\n%s", util.Indent(strings.TrimSpace(utiltest.GenerateRSAPrivateKey(t)), "  ")),
			false,
		},
	}

	for n, testCase := range cases {
		testCase := testCase

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			bar := &Foo{}

			err := yaml.Unmarshal([]byte(testCase.YAML), bar)
			if testCase.Error && err == nil {
				t.Fatalf("Expected error and didn't get any.")
			}

			if !testCase.Error && err != nil {
				t.Fatalf("Didn't expect error, got: %v", err)
			}

			if err == nil && bar.Bar == "" {
				t.Fatalf("Didn't get any error, but field is empty")
			}

			if err != nil && bar.Bar != "" {
				t.Fatalf("Got error and still got some content")
			}
		})
	}
}

func TestParsePrivateKeyPKCS1(t *testing.T) {
	t.Parallel()

	d := fmt.Sprintf("bar: |\n%s", util.Indent(strings.TrimSpace(utiltest.GeneratePKCS1PrivateKey(t)), "  "))

	if err := yaml.Unmarshal([]byte(d), &Foo{}); err != nil {
		t.Fatalf("Parsing valid PKCS1 private key should succeed, got: %v", err)
	}
}

func TestParsePrivateKeyEC(t *testing.T) {
	t.Parallel()

	d := fmt.Sprintf("bar: |\n%s", util.Indent(strings.TrimSpace(utiltest.GenerateECPrivateKey(t)), "  "))

	if err := yaml.Unmarshal([]byte(d), &Foo{}); err != nil {
		t.Fatalf("Parsing valid EC private key should succeed, got: %v", err)
	}
}

func TestParsePrivateKeyBad(t *testing.T) {
	t.Parallel()

	//#nosec G101 // Just bad test data.
	privateKey := `---
bar: |
  -----BEGIN RSA PRIVATE KEY-----
  Zm9vCg==
  -----END RSA PRIVATE KEY-----
`
	if err := yaml.Unmarshal([]byte(privateKey), &Foo{}); err == nil {
		t.Fatalf("Parsing not PEM format should fail")
	}
}

func TestPrivateKeyPickNil(t *testing.T) {
	t.Parallel()

	var c types.PrivateKey

	if c.Pick(types.PrivateKey("bar"), types.PrivateKey("baz")) != "bar" {
		t.Fatalf("First non empty private key should be picked")
	}
}

func TestPrivateKeyPick(t *testing.T) {
	t.Parallel()

	d := types.PrivateKey("foo")

	if d.Pick(types.PrivateKey("baz")) != "foo" {
		t.Fatalf("First non empty private key should be picked")
	}
}
