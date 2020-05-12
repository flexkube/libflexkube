package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
)

func TestClientUnmarshal(t *testing.T) {
	c := &client.Config{
		Server: "foo",
		Token:  "bar",
	}

	e := []interface{}{
		map[string]interface{}{
			"server": "foo",
			"token":  "bar",
		},
	}

	if diff := cmp.Diff(clientUnmarshal(e[0]), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestClientUnmarshalEmpty(t *testing.T) {
	var c *client.Config
	if diff := cmp.Diff(clientUnmarshal(nil), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestClientUnmarshalEmptyBock(t *testing.T) {
	var c *client.Config

	if diff := cmp.Diff(clientUnmarshal(map[string]interface{}{}), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
