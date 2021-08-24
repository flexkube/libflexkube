package host

import (
	"fmt"
	"testing"

	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

// New() tests.
func TestNew(t *testing.T) {
	t.Parallel()

	h := BuildConfig(Host{
		SSHConfig: &ssh.Config{
			Address:  "localhost",
			Password: "foo",
		},
	}, Host{})

	if _, err := h.New(); err != nil {
		t.Fatalf("Built config should be valid, got: %v", err)
	}
}

func TestNewValidate(t *testing.T) {
	t.Parallel()

	h := &Host{}

	if _, err := h.New(); err == nil {
		t.Fatalf("New should validate the configuration")
	}
}

// Validate() tests.
func TestValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Host    *Host
		Message string
		Error   bool
	}{
		{
			&Host{
				SSHConfig:    &ssh.Config{},
				DirectConfig: &direct.Config{},
			},
			"Validate should reject ambiguous configuration",
			true,
		},
		{
			&Host{},
			"Validate should reject empty configuration",
			true,
		},
		{
			&Host{
				SSHConfig: &ssh.Config{},
			},
			"Validate must validate ssh configuration",
			true,
		},
	}

	for n, c := range cases {
		c := c

		t.Run(fmt.Sprintf("%d", n), func(t *testing.T) {
			t.Parallel()

			err := c.Host.Validate()
			if c.Error && err == nil {
				t.Fatalf(c.Message)
			}
			if !c.Error && err != nil {
				t.Errorf(c.Message)
			}
		})
	}
}

// Connect() tests.
func TestConnect(t *testing.T) {
	t.Parallel()

	h := Host{
		DirectConfig: &direct.Config{},
	}

	c, err := h.New()
	if err != nil {
		t.Fatalf("Config should be valid, got: %v", err)
	}

	if _, err := c.Connect(); err != nil {
		t.Fatalf("Direct config should always connect, got: %v", err)
	}
}

// ForwardUnixSocket() tests.
func TestForwardUnixSocket(t *testing.T) {
	t.Parallel()

	h := Host{
		DirectConfig: &direct.Config{},
	}

	c, err := h.New()
	if err != nil {
		t.Fatalf("Config should be valid, got: %v", err)
	}

	hc, err := c.Connect()
	if err != nil {
		t.Fatalf("Direct config should always connect, got: %v", err)
	}

	if _, err := hc.ForwardUnixSocket("unix:///nonexisting"); err != nil {
		t.Fatalf("Forwarding shouldn't fail, got: %v", err)
	}
}

// ForwardTCP() tests.
func TestForwardTCP(t *testing.T) {
	t.Parallel()

	h := Host{
		DirectConfig: &direct.Config{},
	}

	c, err := h.New()
	if err != nil {
		t.Fatalf("Config should be valid, got: %v", err)
	}

	hc, err := c.Connect()
	if err != nil {
		t.Fatalf("Direct config should always connect, got: %v", err)
	}

	if _, err := hc.ForwardTCP("localhost:80"); err != nil {
		t.Fatalf("Forwarding shouldn't fail, got: %v", err)
	}
}

// BuildConfig() tests.
func TestBuildConfigDirectByDefault(t *testing.T) {
	t.Parallel()

	h := BuildConfig(Host{}, Host{})
	if err := h.Validate(); err != nil {
		t.Errorf("Config returned by default should be valid, got: %v", err)
	}

	if h.DirectConfig == nil {
		t.Fatalf("BuildConfig should return direct config by default")
	}
}

func TestBuildConfigFirstPriorityDirect(t *testing.T) {
	t.Parallel()

	c := Host{
		DirectConfig: &direct.Config{},
	}

	d := Host{
		SSHConfig: &ssh.Config{},
	}

	h := BuildConfig(c, d)
	if err := h.Validate(); err != nil {
		t.Errorf("Config returned by default should be valid, got: %v", err)
	}

	if h.SSHConfig != nil {
		t.Fatalf("BuildConfig should not inject SSH configuration from defaults if direct configuration has been requested")
	}
}

func TestBuildConfigFirstPriotitySSH(t *testing.T) {
	t.Parallel()

	c := Host{
		SSHConfig: &ssh.Config{
			Address:  "foo",
			Password: "foo",
		},
	}

	d := Host{
		DirectConfig: &direct.Config{},
	}

	h := BuildConfig(c, d)
	if err := h.Validate(); err != nil {
		t.Errorf("Config returned should be valid, got: %v", err)
	}

	if h.DirectConfig != nil {
		t.Fatalf("BuildConfig should not have direct configuration from defaults if SSH configuration has been requested")
	}
}

func TestBuildConfigSSH(t *testing.T) {
	t.Parallel()

	u := Host{
		SSHConfig: &ssh.Config{
			Address: "foo",
		},
	}

	d := Host{
		SSHConfig: &ssh.Config{
			Port: 33,
		},
	}

	if h := BuildConfig(u, d); h.SSHConfig.Port != 33 || h.SSHConfig.Address != "foo" {
		t.Fatalf("BuildConfig should merge ssh config, got: %+v", h)
	}
}
