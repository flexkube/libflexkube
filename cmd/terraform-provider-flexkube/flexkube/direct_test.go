package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func TestDirectMarshal(t *testing.T) {
	c := direct.Config{}

	e := []interface{}{
		map[string]interface{}{},
	}

	if diff := cmp.Diff(directMarshal(c), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestDirectUnmarshal(t *testing.T) {
	c := &direct.Config{}

	e := []interface{}{
		map[string]interface{}{},
	}

	if diff := cmp.Diff(directUnmarshal(e[0]), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
