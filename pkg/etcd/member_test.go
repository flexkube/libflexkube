package etcd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/etcdserverpb"

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

func TestNewCluster(t *testing.T) {
	m := &member{
		newCluster: true,
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
	m := &member{
		newCluster: false,
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

// peerURLs()
func TestPeerURLs(t *testing.T) {
	m := &member{
		peerAddress: "1.1.1.1",
	}

	e := "https://1.1.1.1:2380"
	if urls := m.peerURLs(); urls[0] != e {
		t.Fatalf("expected %s, got %s", e, urls[0])
	}
}

// forwardEndpoints()
func TestForwardEndpoints(t *testing.T) {
	m := &member{
		peerAddress: "127.0.0.1",
		host: host.Host{
			DirectConfig: &direct.Config{},
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
	m := &member{
		peerAddress: "127.0.0.1",
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	if _, err := m.forwardEndpoints([]string{"127.0.0.1"}); err == nil {
		t.Fatalf("Forwarding bad address should fail")
	}
}

// getID()
func TestGetIDFailToListMembers(t *testing.T) {
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
		name: "foo",
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
	}

	m := &member{
		peerAddress: "foo",
	}

	id, err := m.getID(f)
	if err != nil {
		t.Fatalf("Getting member ID should work, got: %v", err)
	}

	if id != testID {
		t.Fatalf("Member ID should be %d when member is present, got %d", testID, id)
	}
}

// getEtcdClient()
func TestGetEtcdClientNoEndpoints(t *testing.T) {
	m := &member{
		caCertificate: utiltest.GenerateX509Certificate(t),
	}

	if _, err := m.getEtcdClient([]string{}); err == nil {
		t.Fatalf("Creating etcd client with no endpoints should fail")
	}
}

func TestGetEtcdClient(t *testing.T) {
	m := &member{
		peerCertificate: "",
		peerKey:         "",
		caCertificate:   utiltest.GenerateX509Certificate(t),
		host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}

	if _, err := m.getEtcdClient([]string{"foo"}); err != nil {
		t.Fatalf("Creating etcd client should succeed, got: %v", err)
	}
}

const testID = 1

// remove()
func TestRemove(t *testing.T) {
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
		name: "foo",
	}

	if err := m.remove(f); err != nil {
		t.Fatalf("Removing member should work, got: %v", err)
	}
}

func TestRemoveNonExistent(t *testing.T) {
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
			return nil, fmt.Errorf("expected")
		},
	}
	defer f.Close()

	m := &member{
		name: "foo",
	}

	if err := m.remove(f); err == nil {
		t.Fatalf("Removing member should check for removal errors")
	}
}

func TestRemoveGetIDFail(t *testing.T) {
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

// addMember()
func TestAddMember(t *testing.T) {
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

	m := &member{}

	if err := m.add(f); err != nil {
		t.Fatalf("Adding member should work, got: %v", err)
	}
}

func TestAddMemberAlreadyExists(t *testing.T) {
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
		peerAddress: "foo",
	}

	if err := m.add(f); err != nil {
		t.Fatalf("Adding already existing member shouldn't trigger adding, got error: %v", err)
	}
}

func TestAddMemberFail(t *testing.T) {
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

	m := &member{}

	if err := m.add(f); err == nil {
		t.Fatalf("Adding member should check for adding errors")
	}
}

func TestAddGetIDFail(t *testing.T) {
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
