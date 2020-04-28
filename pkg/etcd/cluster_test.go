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
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

// FromYAML() tests.
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

	cluster, err := FromYaml(buf.Bytes())
	if err != nil {
		t.Fatalf("Creating etcd cluster from YAML should succeed, got: %v", err)
	}

	if err := cluster.CheckCurrentState(); err != nil {
		t.Fatalf("Checking current state for empty cluster should work, got: %v", err)
	}

	if c := cluster.Containers(); c == nil {
		t.Fatalf("Containers() should return valid ContainersInterface object")
	}

	if _, err := cluster.StateToYaml(); err != nil {
		t.Fatalf("Dumping cluster state to YAML should work, got: %v", err)
	}
}

// New() tests.
func TestNewValidateFail(t *testing.T) {
	config := &Cluster{}

	if _, err := config.New(); err == nil {
		t.Fatalf("New() should validate cluster configuration and fail on bad configuration")
	}
}

// Validate() tests.
func TestValidateValidateMembers(t *testing.T) {
	config := &Cluster{
		Members: map[string]Member{
			"foo": {},
		},
	}

	if err := config.Validate(); err == nil {
		t.Fatalf("Should validate members")
	}
}

// getExistingEndpoints() tests.
func TestExistingEndpointsNoEndpoints(t *testing.T) {
	c := &cluster{}
	if len(c.getExistingEndpoints()) != 0 {
		t.Fatalf("No endpoints should be returned for empty cluster")
	}
}

func TestExistingEndpoints(t *testing.T) {
	c := &cluster{
		containers: getContainers(t),
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

// firstMember() tests.
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

func getContainers(t *testing.T) container.ContainersInterface {
	cc := &container.Containers{
		PreviousState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	co, err := cc.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	return co
}

func getFakeHostConfiguredContainer() *container.HostConfiguredContainer {
	return &container.HostConfiguredContainer{
		Container: container.Container{
			Config: types.ContainerConfig{
				Name:  "foo",
				Image: "bar",
			},
			Runtime: container.RuntimeConfig{
				Docker: docker.DefaultConfig(),
			},
		},
		Host: host.Host{
			DirectConfig: &direct.Config{},
		},
	}
}

// getClient() tests.
func TestGetClientEmptyCluster(t *testing.T) {
	c := &cluster{}
	if _, err := c.getClient(); err == nil {
		t.Fatalf("Getting client on empty cluster should fail")
	}
}

func TestGetClientForwardFail(t *testing.T) {
	c := &cluster{
		containers: getContainers(t),
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
		containers: getContainers(t),
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

// membersToRemove() tests.
func TestMembersToRemove(t *testing.T) {
	cc := &container.Containers{
		PreviousState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
			"bar": getFakeHostConfiguredContainer(),
		},
		DesiredState: container.ContainersState{
			"bar": getFakeHostConfiguredContainer(),
		},
	}

	co, err := cc.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	c := &cluster{
		containers: co,
	}

	e := []string{"foo"}

	if r := c.membersToRemove(); !reflect.DeepEqual(r, e) {
		t.Fatalf("Expected %+v, got %+v", e, r)
	}
}

// membersToAdd() tests.
func TestMembersToAdd(t *testing.T) {
	cc := &container.Containers{
		PreviousState: container.ContainersState{
			"bar": getFakeHostConfiguredContainer(),
		},
		DesiredState: container.ContainersState{
			"bar": getFakeHostConfiguredContainer(),
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	co, err := cc.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	c := &cluster{
		containers: co,
	}

	e := []string{"foo"}

	if r := c.membersToAdd(); !reflect.DeepEqual(r, e) {
		t.Fatalf("Expected %+v, got %+v", e, r)
	}
}

// updateMembers() tests.
func TestUpdateMembersNoUpdates(t *testing.T) {
	cc := &container.Containers{
		PreviousState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
		DesiredState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	co, err := cc.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	c := &cluster{
		containers: co,
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
		containers: getContainers(t),
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

	if err := c.updateMembers(f); err == nil {
		t.Fatalf("Removing member should fail")
	}
}

func TestUpdateMembersAddMember(t *testing.T) {
	cc := &container.Containers{
		DesiredState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	co, err := cc.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	c := &cluster{
		containers: co,
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

// Deploy() tests.
func TestDeploy(t *testing.T) {
	cc := &container.Containers{
		DesiredState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	co, err := cc.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	c := &cluster{
		containers: co,
		members:    map[string]*member{},
	}

	err = c.Deploy()
	if err == nil {
		t.Fatalf("Deploying bad containers should fail")
	}

	if !strings.Contains(err.Error(), "without knowing current state of the containers") {
		t.Fatalf("Deploying new cluster should not trigger updateMembers and fail on deploying, got: %v", err)
	}
}

func TestDeployUpdateMembers(t *testing.T) {
	cc := &container.Containers{
		PreviousState: container.ContainersState{
			"bar": getFakeHostConfiguredContainer(),
		},
		DesiredState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	co, err := cc.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	c := &cluster{
		containers: co,
		members:    map[string]*member{},
	}

	err = c.Deploy()
	if err == nil {
		t.Fatalf("Deploying should trigger updateMembers and fail")
	}

	if !strings.Contains(err.Error(), "failed getting etcd client") {
		t.Fatalf("Expected failure in client creation, got: %v", err)
	}
}
