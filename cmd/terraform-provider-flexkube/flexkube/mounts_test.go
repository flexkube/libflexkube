package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container/types"
)

func TestMountsMarshal(t *testing.T) {
	c := []types.Mount{
		{
			Source:      "/foo",
			Target:      "/bar",
			Propagation: "bidirectional",
		},
	}

	e := []interface{}{
		map[string]interface{}{
			"source":      "/foo",
			"target":      "/bar",
			"propagation": "bidirectional",
		},
	}

	if diff := cmp.Diff(mountsMarshal(c), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestMountsUnmarshal(t *testing.T) {
	c := []types.Mount{
		{
			Source:      "/foo",
			Target:      "/bar",
			Propagation: "bidirectional",
		},
	}

	e := []interface{}{
		map[string]interface{}{
			"source":      "/foo",
			"target":      "/bar",
			"propagation": "bidirectional",
		},
	}

	if diff := cmp.Diff(mountsUnmarshal(e), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
