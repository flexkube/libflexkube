package docker

import "testing"

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
