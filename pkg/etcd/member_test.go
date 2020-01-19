package etcd

import (
	"testing"

	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/types"
)

const (
	nonEmptyString = "foo"
)

func TestMemberToHostConfiguredContainer(t *testing.T) {
	cert := types.Certificate(utiltest.GenerateX509Certificate(t))
	privateKey := types.PrivateKey(utiltest.GenerateRSAPrivateKey(t))

	kas := &Member{
		Name:              nonEmptyString,
		PeerAddress:       nonEmptyString,
		CACertificate:     cert,
		PeerCertificate:   cert,
		PeerKey:           privateKey,
		ServerCertificate: cert,
		ServerKey:         privateKey,
		Host: host.Host{
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
