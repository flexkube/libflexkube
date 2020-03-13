package flexkube

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/google/go-cmp/cmp"
)

func TestContainerStatusMarshal(t *testing.T) {
	cs := types.ContainerStatus{
		ID:     "foo",
		Status: "running",
	}

	e := []interface{}{
		map[string]interface{}{
			"id":     "foo",
			"status": "running",
		},
	}

	if diff := cmp.Diff(containerStatusMarshal(cs), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestContainerStatusUnmarshal(t *testing.T) {
	cs := &types.ContainerStatus{
		ID:     "foo",
		Status: "running",
	}

	e := []interface{}{
		map[string]interface{}{
			"id":     "foo",
			"status": "running",
		},
	}

	if diff := cmp.Diff(containerStatusUnmarshal(e[0]), cs); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
