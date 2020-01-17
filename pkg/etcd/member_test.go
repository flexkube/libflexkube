package etcd

import (
	"testing"

	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

const (
	nonEmptyString = "foo"
)

func TestMemberToHostConfiguredContainer(t *testing.T) {
	kas := &Member{
		Name:        nonEmptyString,
		PeerAddress: nonEmptyString,
		Host: &host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	o, err := kas.New()
	if err != nil {
		t.Fatalf("new should not return error, got: %v", err)
	}

	hcc, err := o.ToHostConfiguredContainer()
	if err != nil {
		t.Fatalf("Generating HostConfiguredContainer should work, got: %v", err)
	}

	if _, err := hcc.New(); err != nil {
		t.Fatalf("ToHostConfiguredContainer() should generate valid HostConfiguredContainer, got: %v", err)
	}
}
