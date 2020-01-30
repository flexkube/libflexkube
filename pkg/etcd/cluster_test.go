package etcd

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/etcdserverpb"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

// FromYAML()
func TestClusterFromYaml(t *testing.T) {
	c := `
ssh:
  user: "core"
  port: 2222
  password: foo
caCertificate: |
  {{.Certificate}}
members:
  foo:
    peerCertificate: |
      {{.MemberCertificate}}
    peerKey: |
      {{.MemberKey}}
    serverCertificate: |
      {{.MemberCertificate}}
    serverKey: |
      {{.MemberKey}}
    host:
      ssh:
        address: "127.0.0.1"
    peerAddress: 10.0.2.15
    serverAddress: 10.0.2.15
`

	data := struct {
		Certificate       string
		MemberCertificate string
		MemberKey         string
	}{
		strings.TrimSpace(util.Indent(utiltest.GenerateX509Certificate(t), "  ")),
		strings.TrimSpace(util.Indent(utiltest.GenerateX509Certificate(t), "      ")),
		strings.TrimSpace(util.Indent(utiltest.GenerateRSAPrivateKey(t), "      ")),
	}

	var buf bytes.Buffer

	tpl := template.Must(template.New("c").Parse(c))
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to generate config from template: %v", err)
	}

	if _, err := FromYaml(buf.Bytes()); err != nil {
		t.Fatalf("Creating etcd cluster from YAML should succeed, got: %v", err)
	}
}

// getExistingEndpoints()
func TestExistingEndpointsNoEndpoints(t *testing.T) {
	c := &cluster{}
	if len(c.getExistingEndpoints()) != 0 {
		t.Fatalf("No endpoints should be returned for empty cluster")
	}
}

func TestExistingEndpoints(t *testing.T) {
	c := &cluster{
		containers: container.Containers{
			PreviousState: container.ContainersState{
				"foo": &container.HostConfiguredContainer{},
			},
		},
		members: map[string]*member{
			"foo": {
				peerAddress: "1.1.1.1",
			},
		},
	}

	e := []string{"1.1.1.1:2379"}

	if ee := c.getExistingEndpoints(); !reflect.DeepEqual(e, ee) {
		t.Fatalf("Expected %+v, got %+v", e, ee)
	}
}

// firstMember()
func TestFirstMemberNoMembers(t *testing.T) {
	c := &cluster{}

	if _, err := c.firstMember(); err == nil {
		t.Fatalf("Selecting first member on empty cluster should fail")
	}
}

func TestFirstMember(t *testing.T) {
	c := &cluster{
		members: map[string]*member{
			"foo": {},
		},
	}

	f, err := c.firstMember()
	if err != nil {
		t.Fatalf("Selecting first member should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(f, c.members["foo"]) {
		t.Fatalf("Expected %+v, got %+v", c.members["foo"], f)
	}
}

// getClient()
func TestGetClientEmptyCluster(t *testing.T) {
	c := &cluster{}
	if _, err := c.getClient(); err == nil {
		t.Fatalf("Getting client on empty cluster should fail")
	}
}

func TestGetClientForwardFail(t *testing.T) {
	c := &cluster{
		members: map[string]*member{
			"foo": {
				host: host.Host{
					SSHConfig: ssh.BuildConfig(&ssh.Config{
						Address:           "localhost",
						Password:          "foo",
						ConnectionTimeout: "1ms",
						RetryTimeout:      "1ms",
						RetryInterval:     "1ms",
					}, nil),
				},
			},
		},
	}

	if _, err := c.getClient(); err == nil {
		t.Fatalf("Getting client on empty cluster should fail")
	}
}

func TestGetClient(t *testing.T) {
	c := &cluster{
		containers: container.Containers{
			PreviousState: container.ContainersState{
				"foo": &container.HostConfiguredContainer{},
			},
		},
		members: map[string]*member{
			"foo": {
				peerCertificate: "",
				peerKey:         "",
				caCertificate:   utiltest.GenerateX509Certificate(t),
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
	}

	if _, err := c.getClient(); err != nil {
		t.Fatalf("Getting client should succeed, got: %v", err)
	}
}

// membersToRemove()
func TestMembersToRemove(t *testing.T) {
	c := &cluster{
		containers: container.Containers{
			PreviousState: container.ContainersState{
				"foo": &container.HostConfiguredContainer{},
				"bar": &container.HostConfiguredContainer{},
			},
			DesiredState: container.ContainersState{
				"bar": &container.HostConfiguredContainer{},
			},
		},
	}

	e := []string{"foo"}

	if r := c.membersToRemove(); !reflect.DeepEqual(r, e) {
		t.Fatalf("Expected %+v, got %+v", e, r)
	}
}

// membersToAdd()
func TestMembersToAdd(t *testing.T) {
	c := &cluster{
		containers: container.Containers{
			PreviousState: container.ContainersState{
				"bar": &container.HostConfiguredContainer{},
			},
			DesiredState: container.ContainersState{
				"bar": &container.HostConfiguredContainer{},
				"foo": &container.HostConfiguredContainer{},
			},
		},
	}

	e := []string{"foo"}

	if r := c.membersToAdd(); !reflect.DeepEqual(r, e) {
		t.Fatalf("Expected %+v, got %+v", e, r)
	}
}

// updateMembers()
func TestUpdateMembersNoUpdates(t *testing.T) {
	c := &cluster{
		containers: container.Containers{
			PreviousState: container.ContainersState{
				"foo": &container.HostConfiguredContainer{},
			},
			DesiredState: container.ContainersState{
				"foo": &container.HostConfiguredContainer{},
			},
		},
		members: map[string]*member{
			"foo": {
				peerCertificate: "",
				peerKey:         "",
				caCertificate:   utiltest.GenerateX509Certificate(t),
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
	}

	f := &fakeClient{}

	if err := c.updateMembers(f); err != nil {
		t.Fatalf("Updating members without any pending updates should succeed, got: %v", err)
	}
}

func TestUpdateMembersRemoveMember(t *testing.T) {
	c := &cluster{
		containers: container.Containers{
			PreviousState: container.ContainersState{
				"foo": &container.HostConfiguredContainer{},
			},
		},
		members: map[string]*member{
			"foo": {
				name:            "foo",
				peerCertificate: "",
				peerKey:         "",
				caCertificate:   utiltest.GenerateX509Certificate(t),
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
	}

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

	if err := c.updateMembers(f); err == nil {
		t.Fatalf("Removing member should fail")
	}
}

func TestUpdateMembersAddMember(t *testing.T) {
	c := &cluster{
		containers: container.Containers{
			DesiredState: container.ContainersState{
				"foo": &container.HostConfiguredContainer{},
			},
		},
		members: map[string]*member{
			"foo": {
				name:            "foo",
				peerCertificate: "",
				peerKey:         "",
				caCertificate:   utiltest.GenerateX509Certificate(t),
				host: host.Host{
					DirectConfig: &direct.Config{},
				},
			},
		},
	}

	f := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{},
			}, nil
		},
		memberAddF: func(context context.Context, peerURLs []string) (*clientv3.MemberAddResponse, error) {
			return nil, fmt.Errorf("expected")
		},
	}

	if err := c.updateMembers(f); err == nil {
		t.Fatalf("Adding member should fail")
	}
}

// Deploy()
func TestDeploy(t *testing.T) {
	c := &cluster{
		containers: container.Containers{
			DesiredState: container.ContainersState{
				"foo": &container.HostConfiguredContainer{},
			},
		},
		members: map[string]*member{},
	}

	err := c.Deploy()
	if err == nil {
		t.Fatalf("Deploying bad containers should fail")
	}

	if !strings.Contains(err.Error(), "failed to validate containers configuration") {
		t.Fatalf("Deploying new cluster should not trigger updateMembers and fail on container configuration, got: %v", err)
	}
}

func TestDeployUpdateMembers(t *testing.T) {
	c := &cluster{
		containers: container.Containers{
			PreviousState: container.ContainersState{
				"bar": &container.HostConfiguredContainer{},
			},
			DesiredState: container.ContainersState{
				"foo": &container.HostConfiguredContainer{},
			},
		},
		members: map[string]*member{},
	}

	err := c.Deploy()
	if err == nil {
		t.Fatalf("Deploying should trigger updateMembers and fail")
	}

	if !strings.Contains(err.Error(), "failed getting etcd client") {
		t.Fatalf("Expected failure in client creation, got: %v", err)
	}
}
