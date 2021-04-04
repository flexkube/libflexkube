package etcd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
)

func TestNewCluster(t *testing.T) {
	t.Parallel()

	m := &member{
		config: &MemberConfig{
			NewCluster: true,
		},
	}

	hcc, err := m.ToHostConfiguredContainer()
	if err != nil {
		t.Fatalf("Creating host configured container should succeed, got: %v", err)
	}

	flag := false

	for _, f := range hcc.Container.Config.Args {
		if strings.Contains(f, "--initial-cluster-token") {
			flag = true

			break
		}
	}

	if !flag {
		t.Fatalf("Member of new cluster should have --initial-cluster-token flag set")
	}
}

func TestExistingCluster(t *testing.T) {
	t.Parallel()

	m := &member{
		config: &MemberConfig{
			NewCluster: false,
		},
	}

	hcc, err := m.ToHostConfiguredContainer()
	if err != nil {
		t.Fatalf("Creating host configured container should succeed, got: %v", err)
	}

	flag := false

	for _, f := range hcc.Container.Config.Args {
		if strings.Contains(f, "--initial-cluster-state=existing") {
			flag = true

			break
		}
	}

	if !flag {
		t.Fatalf("New member of existing cluster should have --initial-cluster-state=existing flag set")
	}
}

// peerURLs() tests.
func TestPeerURLs(t *testing.T) {
	t.Parallel()

	m := &member{
		config: &MemberConfig{
			PeerAddress: "1.1.1.1",
		},
	}

	e := "https://1.1.1.1:2380" //nolint:ifshort
	if urls := m.peerURLs(); urls[0] != e {
		t.Fatalf("expected %s, got %s", e, urls[0])
	}
}

// forwardEndpoints() tests.
func TestForwardEndpoints(t *testing.T) {
	t.Parallel()

	m := &member{
		config: &MemberConfig{
			PeerAddress: "127.0.0.1",
			Host: host.Host{
				DirectConfig: &direct.Config{},
			},
		},
	}

	fe, err := m.forwardEndpoints([]string{"127.0.0.1:2379"})
	if err != nil {
		t.Fatalf("Forwarding should succeed, got: %v", err)
	}

	if l := len(fe); l != testID {
		t.Fatalf("Should get exactly one forwarded endpoint, got %d", l)
	}
}

func TestForwardEndpointsFail(t *testing.T) {
	t.Parallel()

	m := &member{
		config: &MemberConfig{
			PeerAddress: "127.0.0.1",
			Host: host.Host{
				DirectConfig: &direct.Config{},
			},
		},
	}

	if _, err := m.forwardEndpoints([]string{"127.0.0.1"}); err == nil {
		t.Fatalf("Forwarding bad address should fail")
	}
}

// getID() tests.
func TestGetIDFailToListMembers(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return nil, fmt.Errorf("expected")
		},
	}

	m := &member{}

	if _, err := m.getID(f); err == nil {
		t.Fatalf("Should return error when listing members fails")
	}
}

func TestGetIDNotFound(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return &clientv3.MemberListResponse{}, nil
		},
	}

	m := &member{}

	id, err := m.getID(f)
	if err != nil {
		t.Fatalf("Getting member ID should work, got: %v", err)
	}

	if id != 0 {
		t.Fatalf("Member ID should be 0 when not found, got %d", id)
	}
}

func TestGetIDByName(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						Name: "etcd-foo",
						ID:   testID,
					},
				},
			}, nil
		},
	}

	m := &member{
		config: &MemberConfig{
			Name: "etcd-foo",
		},
	}

	id, err := m.getID(f)
	if err != nil {
		t.Fatalf("Getting member ID should work, got: %v", err)
	}

	if id != testID {
		t.Fatalf("Member ID should be %d when member is present, got %d", testID, id)
	}
}

func TestGetIDByPeerURL(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						Name:     "foo",
						ID:       testID,
						PeerURLs: []string{"https://foo:2380"},
					},
				},
			}, nil
		},
	}

	m := &member{
		config: &MemberConfig{
			PeerAddress: "foo",
		},
	}

	id, err := m.getID(f)
	if err != nil {
		t.Fatalf("Getting member ID should work, got: %v", err)
	}

	if id != testID {
		t.Fatalf("Member ID should be %d when member is present, got %d", testID, id)
	}
}

// getEtcdClient() tests.
func TestGetEtcdClientNoEndpoints(t *testing.T) {
	t.Parallel()

	m := &member{
		config: &MemberConfig{
			CACertificate: utiltest.GenerateX509Certificate(t),
		},
	}

	if _, err := m.getEtcdClient([]string{}); err == nil {
		t.Fatalf("Creating etcd client with no endpoints should fail")
	}
}

func TestGetEtcdClient(t *testing.T) {
	t.Parallel()

	m := &member{
		config: &MemberConfig{
			PeerCertificate: "",
			PeerKey:         "",
			CACertificate:   utiltest.GenerateX509Certificate(t),
			Host: host.Host{
				DirectConfig: &direct.Config{},
			},
		},
	}

	if _, err := m.getEtcdClient([]string{"foo"}); err != nil {
		t.Fatalf("Creating etcd client should succeed, got: %v", err)
	}
}

const testID = 1

// remove() tests.
func TestRemove(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						Name:     "etcd-foo",
						ID:       testID,
						PeerURLs: []string{"foo"},
					},
				},
			}, nil
		},
		memberRemoveF: func(context context.Context, id uint64) (*clientv3.MemberRemoveResponse, error) {
			return nil, nil
		},
	}

	m := &member{
		config: &MemberConfig{
			Name: "foo",
		},
	}

	if err := m.remove(f); err != nil {
		t.Fatalf("Removing member should work, got: %v", err)
	}
}

func TestRemoveNonExistent(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{},
			}, nil
		},
		memberRemoveF: func(context context.Context, id uint64) (*clientv3.MemberRemoveResponse, error) {
			return nil, nil
		},
	}

	m := &member{}

	if err := m.remove(f); err != nil {
		t.Fatalf("Removing non-existing member shouldn't return error, got: %v", err)
	}
}

func TestRemoveMemberFail(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						Name:     "foo",
						ID:       testID,
						PeerURLs: []string{"foo"},
					},
				},
			}, nil
		},
		memberRemoveF: func(context context.Context, id uint64) (*clientv3.MemberRemoveResponse, error) {
			return nil, fmt.Errorf("expected")
		},
	}

	m := &member{
		config: &MemberConfig{
			Name: "foo",
		},
	}

	if err := m.remove(f); err == nil {
		t.Fatalf("Removing member should check for removal errors")
	}

	if err := f.Close(); err != nil {
		t.Logf("failed closing etcd client: %v", err)
	}
}

func TestRemoveGetIDFail(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return nil, fmt.Errorf("expected")
		},
	}

	m := &member{}

	if err := m.remove(f); err == nil {
		t.Fatalf("Removing member should fail, when getting member id fails")
	}
}

// addMember() tests.
func TestAddMember(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						Name:     "etcd-foo",
						ID:       testID,
						PeerURLs: []string{"foo"},
					},
				},
			}, nil
		},
		memberAddF: func(context context.Context, peerURLs []string) (*clientv3.MemberAddResponse, error) {
			return nil, nil
		},
	}

	m := &member{
		config: &MemberConfig{},
	}

	if err := m.add(f); err != nil {
		t.Fatalf("Adding member should work, got: %v", err)
	}
}

func TestAddMemberAlreadyExists(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						Name:     "etcd-foo",
						ID:       testID,
						PeerURLs: []string{"https://foo:2380"},
					},
				},
			}, nil
		},
		memberAddF: func(context context.Context, peerURLs []string) (*clientv3.MemberAddResponse, error) {
			return nil, fmt.Errorf("expected")
		},
	}

	m := &member{
		config: &MemberConfig{
			PeerAddress: "foo",
		},
	}

	if err := m.add(f); err != nil {
		t.Fatalf("Adding already existing member shouldn't trigger adding, got error: %v", err)
	}
}

func TestAddMemberFail(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						Name:     "etcd-foo",
						ID:       testID,
						PeerURLs: []string{"foo"},
					},
				},
			}, nil
		},
		memberAddF: func(context context.Context, peerURLs []string) (*clientv3.MemberAddResponse, error) {
			return nil, fmt.Errorf("expected")
		},
	}

	m := &member{
		config: &MemberConfig{},
	}

	if err := m.add(f); err == nil {
		t.Fatalf("Adding member should check for adding errors")
	}
}

func TestAddGetIDFail(t *testing.T) {
	t.Parallel()

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return nil, fmt.Errorf("expected")
		},
	}

	m := &member{}

	if err := m.add(f); err == nil {
		t.Fatalf("Adding member should fail, when getting member id fails")
	}
}
