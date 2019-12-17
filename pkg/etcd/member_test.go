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

	// TODO grab an object and perform some validation on it?
	o.ToHostConfiguredContainer()
}
