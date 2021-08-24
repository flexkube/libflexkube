package apiloadbalancer

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func TestToHostConfiguredContainer(t *testing.T) {
	t.Parallel()

	kk := &APILoadBalancer{
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
		Servers:     []string{"localhost:9090"},
		BindAddress: "0.0.0.0:6434",
	}

	k, err := kk.New()
	if err != nil {
		t.Fatalf("Creating new api loadbalancer should succeed, got: %v", err)
	}

	hcc, err := k.ToHostConfiguredContainer()
	if err != nil {
		t.Fatalf("Generating HostConfiguredContainer should work, got: %v", err)
	}

	if _, err := hcc.New(); err != nil {
		t.Fatalf("Should produce valid HostConfiguredContainer, got: %v", err)
	}

	if hcc.Container.Config.User == "" {
		t.Fatalf("HostConfiguredContainer should have user set")
	}
}

// Validate() tests.
func TestValidateRequireServers(t *testing.T) {
	t.Parallel()

	kk := &APILoadBalancer{
		BindAddress: "0.0.0.0:6434",
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	if err := kk.Validate(); err == nil {
		t.Fatalf("Validate should require at least one server to be defined")
	}
}

func TestValidateRequireBindAddress(t *testing.T) {
	t.Parallel()

	kk := &APILoadBalancer{
		Servers: []string{"foo"},
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	if err := kk.Validate(); err == nil {
		t.Fatalf("Validate should require at least one server to be defined")
	}
}

// New() tests.
func TestNewValidate(t *testing.T) {
	t.Parallel()

	kk := &APILoadBalancer{
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	if _, err := kk.New(); err == nil {
		t.Fatalf("New should validate configuration before creating object")
	}
}
