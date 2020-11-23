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
	t.Parallel()

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

func validMember(t *testing.T) *etcd.Member {
	cert := utiltest.GenerateX509Certificate(t)
	privateKey := utiltest.GenerateRSAPrivateKey(t)

	return &etcd.Member{
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
}

// Validate() tests.
//
//nolint:funlen
func TestValidate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		mutator     func(m *etcd.Member) *etcd.Member
		expectError bool
	}{
		"valid": {
			func(m *etcd.Member) *etcd.Member { return m },
			false,
		},
		"peer address": {
			func(m *etcd.Member) *etcd.Member {
				m.PeerAddress = ""

				return m
			},
			true,
		},
		"member name": {
			func(m *etcd.Member) *etcd.Member {
				m.Name = ""

				return m
			},
			true,
		},
		"CA certificate": {
			func(m *etcd.Member) *etcd.Member {
				m.CACertificate = nonEmptyString

				return m
			},
			true,
		},
		"peer certificate": {
			func(m *etcd.Member) *etcd.Member {
				m.PeerCertificate = nonEmptyString

				return m
			},
			true,
		},
		"server certificate": {
			func(m *etcd.Member) *etcd.Member {
				m.ServerCertificate = nonEmptyString

				return m
			},
			true,
		},
		"peer key": {
			func(m *etcd.Member) *etcd.Member {
				m.PeerKey = nonEmptyString

				return m
			},
			true,
		},
		"server key": {
			func(m *etcd.Member) *etcd.Member {
				m.ServerKey = nonEmptyString

				return m
			},
			true,
		},
		"bad host": {
			func(m *etcd.Member) *etcd.Member {
				m.Host.DirectConfig = nil

				return m
			},
			true,
		},
	}

	for c, p := range cases {
		p := p

		t.Run(c, func(t *testing.T) {
			t.Parallel()

			m := p.mutator(validMember(t))
			err := m.Validate()

			if p.expectError && err == nil {
				t.Fatalf("expected error")
			}

			if !p.expectError && err != nil {
				t.Fatalf("didn't expect error, got: %v", err)
			}
		})
	}
}
