package types

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
)

func TestCertificateParse(t *testing.T) {
	t.Parallel()

	type Foo struct {
		Bar Certificate `json:"bar"`
	}

	cases := map[string]struct {
		YAML  string
		Error bool
	}{
		"bad": {
			"bar: doh",
			true,
		},
		"good": {
			fmt.Sprintf("bar: |\n%s", util.Indent(strings.TrimSpace(utiltest.GenerateX509Certificate(t)), "  ")),
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

func TestCertificatePickNil(t *testing.T) {
	t.Parallel()

	var c Certificate

	d := Certificate("bar")
	e := Certificate("baz")

	if c.Pick(d, e) != "bar" {
		t.Fatalf("first non empty certificate should be picked")
	}
}

func TestCertificatePick(t *testing.T) {
	t.Parallel()

	d := Certificate("foo")
	e := Certificate("baz")

	if d.Pick(e) != "foo" {
		t.Fatalf("first non empty certificate should be picked")
	}
}
