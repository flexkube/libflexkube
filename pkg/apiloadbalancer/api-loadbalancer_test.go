package apiloadbalancer

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func TestToHostConfiguredContainer(t *testing.T) {
	kk := &APILoadBalancer{
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
		Servers:            []string{"localhost:9090"},
		MetricsBindAddress: "0.0.0.0:6443",
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
		t.Fatalf("should produce valid HostConfiguredContainer, got: %v", err)
	}
}
