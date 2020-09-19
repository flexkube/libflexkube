package etcd_test

import (
	"testing"

	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/etcd"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

const (
	nonEmptyString = "foo"
)

func TestMemberToHostConfiguredContainer(t *testing.T) {
	cert := utiltest.GenerateX509Certificate(t)
	privateKey := utiltest.GenerateRSAPrivateKey(t)

	kas := &etcd.Member{
		Name:              nonEmptyString,
		PeerAddress:       nonEmptyString,
		CACertificate:     cert,
		PeerCertificate:   cert,
		PeerKey:           privateKey,
		ServerCertificate: cert,
		ServerKey:         privateKey,
		Image:             defaults.EtcdImage,
		PeerCertAllowedCN: nonEmptyString,
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

// Validate() tests.
func TestValidateNoName(t *testing.T) {
	m := &etcd.Member{}

	if err := m.Validate(); err == nil {
		t.Fatalf("Validate() should reject members with empty name")
	}
}
