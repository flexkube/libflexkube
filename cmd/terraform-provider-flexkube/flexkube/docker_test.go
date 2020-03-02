package flexkube

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
)

func TestDockerMarshal(t *testing.T) {
	c := docker.Config{
		Host: "unix:///run/docker.sock",
	}

	e := []interface{}{
		map[string]interface{}{
			"host": "unix:///run/docker.sock",
		},
	}

	if diff := cmp.Diff(dockerMarshal(c), e); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestDockerUnmarshal(t *testing.T) {
	c := &docker.Config{
		Host: "unix:///run/docker.sock",
	}

	e := []interface{}{
		map[string]interface{}{
			"host": "unix:///run/docker.sock",
		},
	}

	if diff := cmp.Diff(dockerUnmarshal(e[0]), c); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestDockerUnmarshalEmpty(t *testing.T) {
	if diff := cmp.Diff(dockerUnmarshal(nil), docker.DefaultConfig()); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}

func TestDockerUnmarshalEmptyBock(t *testing.T) {
	if diff := cmp.Diff(dockerUnmarshal(map[string]interface{}{}), docker.DefaultConfig()); diff != "" {
		t.Errorf("Unexpected diff:\n%s", diff)
	}
}
