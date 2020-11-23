package types

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
)

type Foo struct {
	Bar PrivateKey `json:"bar"`
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

	for n, c := range cases {
		c := c

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			bar := &Foo{}

			err := yaml.Unmarshal([]byte(c.YAML), bar)

			if c.Error && err == nil {
				t.Fatalf("Expected error and didn't get any.")
			}

			if !c.Error && err != nil {
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

	if err := parsePrivateKey([]byte("notpem")); err == nil {
		t.Fatalf("parsing not PEM format should fail")
	}
}

func TestPrivateKeyPickNil(t *testing.T) {
	t.Parallel()

	var c PrivateKey

	d := PrivateKey("bar")
	e := PrivateKey("baz")

	if c.Pick(d, e) != "bar" {
		t.Fatalf("first non empty private key should be picked")
	}
}

func TestPrivateKeyPick(t *testing.T) {
	t.Parallel()

	d := PrivateKey("foo")
	e := PrivateKey("baz")

	if d.Pick(e) != "foo" {
		t.Fatalf("first non empty private key should be picked")
	}
}
