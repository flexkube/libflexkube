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

func TestNewClientWithHost(t *testing.T) {
	config := &Config{
		Host: "unix:///foo.sock",
	}
	c, err := config.New()
	if err != nil {
		t.Fatalf("Creating new docker client should work, got: %s", err)
	}
	if dh := (c.(*docker)).cli.DaemonHost(); dh != config.Host {
		t.Fatalf("Client with host set should have '%s' as host, got: '%s'", config.Host, dh)
	}
}
