package docker

import (
	"context"
	"fmt"
	"testing"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/errdefs"
)

// New()
func TestNewClient(t *testing.T) {
	// TODO does this kind of simple tests make sense? Integration tests calls the same thing
	// anyway. Or maybe we should simply skip error checking in itegration tests to simplify them?
	c := &Config{}
	if _, err := c.New(); err != nil {
		t.Errorf("Creating new docker client should work, got: %s", err)
	}
}

// getDockerClient()
func TestNewClientWithHost(t *testing.T) {
	config := &Config{
		Host: "unix:///foo.sock",
	}

	c, err := config.getDockerClient()
	if err != nil {
		t.Fatalf("Creating new docker client should work, got: %s", err)
	}

	if dh := c.DaemonHost(); dh != config.Host {
		t.Fatalf("Client with host set should have '%s' as host, got: '%s'", config.Host, dh)
	}
}

// sanitizeImageName()
func TestSanitizeImageName(t *testing.T) {
	e := "foo:latest"

	if g := sanitizeImageName("foo"); g != e {
		t.Fatalf("Expected '%s', got '%s'", e, g)
	}
}

func TestSanitizeImageNameWithTag(t *testing.T) {
	e := "foo:v0.1.0"

	if g := sanitizeImageName(e); g != e {
		t.Fatalf("Expected '%s', got '%s'", e, g)
	}
}

// Status()
func TestStatus(t *testing.T) {
	es := "running"

	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			ContainerInspectF: func(ctx context.Context, id string) (dockertypes.ContainerJSON, error) {
				return dockertypes.ContainerJSON{
					ContainerJSONBase: &dockertypes.ContainerJSONBase{
						State: &dockertypes.ContainerState{
							Status: es,
						},
					},
				}, nil
			},
		},
	}

	s, err := d.Status("foo")
	if err != nil {
		t.Fatalf("Checking for status should succeed, got: %v", err)
	}

	if s.ID == "" {
		t.Fatalf("ID in status of existing container should not be empty")
	}

	if s.Status != es {
		t.Fatalf("Received status should be %s, got %s", es, s.Status)
	}
}

func TestStatusNotFound(t *testing.T) {
	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			ContainerInspectF: func(ctx context.Context, id string) (dockertypes.ContainerJSON, error) {
				return dockertypes.ContainerJSON{}, errdefs.NotFound(fmt.Errorf("not found"))
			},
		},
	}

	s, err := d.Status("foo")
	if err != nil {
		t.Fatalf("Checking for status should succeed, got: %v", err)
	}

	if s.ID != "" {
		t.Fatalf("ID in status of non-existing container should be empty")
	}
}

func TestStatusRuntimeError(t *testing.T) {
	d := &docker{
		ctx: context.Background(),
		cli: &FakeClient{
			ContainerInspectF: func(ctx context.Context, id string) (dockertypes.ContainerJSON, error) {
				return dockertypes.ContainerJSON{}, fmt.Errorf("can't check status of container")
			},
		},
	}

	if _, err := d.Status("foo"); err == nil {
		t.Fatalf("Checking for status should fail")
	}
}
