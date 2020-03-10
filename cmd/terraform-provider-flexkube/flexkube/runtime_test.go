package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
)

func TestRuntimeMarshal(t *testing.T) {
	c := container.RuntimeConfig{
		Docker: docker.DefaultConfig(),
	}

	e := []interface{}{
		map[string]interface{}{
			"docker": []interface{}{
				map[string]interface{}{
					"host": "unix:///var/run/docker.sock",
				},
			},
		},
	}

	if diff := cmp.Diff(runtimeMarshal(c), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestRuntimeUnmarshal(t *testing.T) {
	c := container.RuntimeConfig{
		Docker: docker.DefaultConfig(),
	}

	e := []interface{}{
		map[string]interface{}{
			"docker": []interface{}{
				map[string]interface{}{
					"host": "unix:///var/run/docker.sock",
				},
			},
		},
	}

	if diff := cmp.Diff(runtimeUnmarshal(e[0]), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestRuntimeUnmarshalEmpty(t *testing.T) {
	rc := container.RuntimeConfig{
		Docker: docker.DefaultConfig(),
	}

	if diff := cmp.Diff(runtimeUnmarshal(nil), rc); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestRuntimeUnmarshalEmptyBock(t *testing.T) {
	rc := container.RuntimeConfig{
		Docker: docker.DefaultConfig(),
	}

	if diff := cmp.Diff(runtimeUnmarshal(map[string]interface{}{}), rc); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
