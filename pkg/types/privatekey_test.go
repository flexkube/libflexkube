package types

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
)

func TestPrivateKeyParse(t *testing.T) {
	type Foo struct {
		Bar PrivateKey `json:"bar"`
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
			fmt.Sprintf("bar: |\n%s", util.Indent(strings.TrimSpace(utiltest.GenerateRSAPrivateKey(t)), "  ")),
			false,
		},
	}

	for n, c := range cases {
		c := c

		t.Run(n, func(t *testing.T) {
			bar := &Foo{}
			if err := yaml.Unmarshal([]byte(c.YAML), bar); !c.Error && err != nil {
				t.Fatalf("Didn't expect error, got: %v", err)
			}
		})
	}
}
