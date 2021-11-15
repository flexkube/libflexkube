package etcd

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/internal/utiltest"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
	"github.com/flexkube/libflexkube/pkg/pki"
)

const yamlConfigTemplate = `
ssh:
  user: "core"
  port: 2222
  password: foo
caCertificate: |
  {{.Certificate}}
extraMounts:
- source: /foo
  destination: /bar
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

// FromYAML() tests.
func TestClusterFromYaml(t *testing.T) {
	t.Parallel()

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

	tpl := template.Must(template.New("c").Parse(yamlConfigTemplate))
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
	t.Parallel()

	config := &Cluster{}

	if _, err := config.New(); err == nil {
		t.Fatalf("New() should validate cluster configuration and fail on bad configuration")
	}
}

// Validate() tests.
func TestValidateValidateMembers(t *testing.T) {
	t.Parallel()

	config := &Cluster{
		Members: map[string]MemberConfig{
			"foo": {},
		},
	}

	if err := config.Validate(); err == nil {
		t.Fatalf("Should validate members")
	}
}

func TestValidateValidatePass(t *testing.T) {
	t.Parallel()

	cert := utiltest.GenerateX509Certificate(t)
	key := utiltest.GenerateRSAPrivateKey(t)

	config := &Cluster{
		Members: map[string]MemberConfig{
			"foo": {
				PeerCertificate:   cert,
				PeerKey:           key,
				ServerCertificate: cert,
				ServerKey:         key,
				PeerAddress:       "1",
				CACertificate:     cert,
			},
		},
	}

	if err := config.Validate(); err != nil {
		t.Fatalf("Valid configuration should pass, got: %v", err)
	}
}

func TestValidateValidateBadCACertificate(t *testing.T) {
	t.Parallel()

	cert := utiltest.GenerateX509Certificate(t)
	key := utiltest.GenerateRSAPrivateKey(t)

	config := &Cluster{
		CACertificate: "doh",
		Members: map[string]MemberConfig{
			"foo": {
				PeerCertificate:   cert,
				PeerKey:           key,
				ServerCertificate: cert,
				ServerKey:         key,
				PeerAddress:       "1",
				CACertificate:     cert,
			},
		},
	}

	if err := config.Validate(); err == nil {
		t.Fatalf("Validation with bad CA certificate should fail")
	}
}

// getExistingEndpoints() tests.
func TestExistingEndpointsNoEndpoints(t *testing.T) {
	t.Parallel()

	testCluster := &cluster{}
	if len(testCluster.getExistingEndpoints()) != 0 {
		t.Fatalf("No endpoints should be returned for empty cluster")
	}
}

func TestExistingEndpoints(t *testing.T) {
	t.Parallel()

	testCluster := &cluster{
		containers: getContainers(t),
		members: map[string]Member{
			"foo": &member{
				config: &MemberConfig{
					PeerAddress: "1.1.1.1",
				},
			},
		},
	}

	e := []string{"1.1.1.1:2379"} //nolint:ifshort // Declare 2 variables in if statement is not common.

	if ee := testCluster.getExistingEndpoints(); !reflect.DeepEqual(e, ee) {
		t.Fatalf("Expected %+v, got %+v", e, ee)
	}
}

// firstMember() tests.
func TestFirstMemberNoMembers(t *testing.T) {
	t.Parallel()

	testCluster := &cluster{}

	if _, err := testCluster.firstMember(); err == nil {
		t.Fatalf("Selecting first member on empty cluster should fail")
	}
}

func TestFirstMember(t *testing.T) {
	t.Parallel()

	testCluster := &cluster{
		members: map[string]Member{
			"foo": &member{},
		},
	}

	firstMember, err := testCluster.firstMember()
	if err != nil {
		t.Fatalf("Selecting first member should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(firstMember, testCluster.members["foo"]) {
		t.Fatalf("Expected %+v, got %+v", testCluster.members["foo"], firstMember)
	}
}

func getContainers(t *testing.T) container.ContainersInterface {
	t.Helper()

	testContainersConfig := &container.Containers{
		PreviousState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	testContainers, err := testContainersConfig.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	return testContainers
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
	t.Parallel()

	testCluster := &cluster{}
	if _, err := testCluster.getClient(); err == nil {
		t.Fatalf("Getting client on empty cluster should fail")
	}
}

func TestGetClientForwardFail(t *testing.T) {
	t.Parallel()

	testCluster := &cluster{
		containers: getContainers(t),
		members: map[string]Member{
			"foo": &member{
				config: &MemberConfig{
					Host: host.Host{
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
		},
	}

	if _, err := testCluster.getClient(); err == nil {
		t.Fatalf("Getting client on empty cluster should fail")
	}
}

func TestGetClient(t *testing.T) {
	t.Parallel()

	testCluster := &cluster{
		containers: getContainers(t),
		members: map[string]Member{
			"foo": &member{
				config: &MemberConfig{
					PeerCertificate: "",
					PeerKey:         "",
					CACertificate:   utiltest.GenerateX509Certificate(t),
					Host: host.Host{
						DirectConfig: &direct.Config{},
					},
				},
			},
		},
	}

	if _, err := testCluster.getClient(); err != nil {
		t.Fatalf("Getting client should succeed, got: %v", err)
	}
}

// membersToRemove() tests.
func TestMembersToRemove(t *testing.T) {
	t.Parallel()

	testContainersConfig := &container.Containers{
		PreviousState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
			"bar": getFakeHostConfiguredContainer(),
		},
		DesiredState: container.ContainersState{
			"bar": getFakeHostConfiguredContainer(),
		},
	}

	testContainers, err := testContainersConfig.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	testCluster := &cluster{
		containers: testContainers,
	}

	e := []string{"foo"} //nolint:ifshort // Declare 2 variables in if statement is not common.

	if r := testCluster.membersToRemove(); !reflect.DeepEqual(r, e) {
		t.Fatalf("Expected %+v, got %+v", e, r)
	}
}

// membersToAdd() tests.
func TestMembersToAdd(t *testing.T) {
	t.Parallel()

	testContainersConfig := &container.Containers{
		PreviousState: container.ContainersState{
			"bar": getFakeHostConfiguredContainer(),
		},
		DesiredState: container.ContainersState{
			"bar": getFakeHostConfiguredContainer(),
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	testContainers, err := testContainersConfig.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	testCluster := &cluster{
		containers: testContainers,
	}

	e := []string{"foo"} //nolint:ifshort // Declare 2 variables in if statement is not common.

	if r := testCluster.membersToAdd(); !reflect.DeepEqual(r, e) {
		t.Fatalf("Expected %+v, got %+v", e, r)
	}
}

// updateMembers() tests.
func TestUpdateMembersNoUpdates(t *testing.T) {
	t.Parallel()

	testContainersConfig := &container.Containers{
		PreviousState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
		DesiredState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	testContainers, err := testContainersConfig.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	testCluster := &cluster{
		containers: testContainers,
		members: map[string]Member{
			"foo": &member{
				config: &MemberConfig{
					PeerCertificate: "",
					PeerKey:         "",
					CACertificate:   utiltest.GenerateX509Certificate(t),
					Host: host.Host{
						DirectConfig: &direct.Config{},
					},
				},
			},
		},
	}

	f := &fakeClient{}

	if err := testCluster.updateMembers(f); err != nil {
		t.Fatalf("Updating members without any pending updates should succeed, got: %v", err)
	}
}

func TestUpdateMembersRemoveMember(t *testing.T) {
	t.Parallel()

	testCluster := &cluster{
		containers: getContainers(t),
		members: map[string]Member{
			"foo": &member{
				config: &MemberConfig{
					Name:            "foo",
					PeerCertificate: "",
					PeerKey:         "",
					CACertificate:   utiltest.GenerateX509Certificate(t),
					Host: host.Host{
						DirectConfig: &direct.Config{},
					},
				},
			},
		},
	}

	testClient := &fakeClient{
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

	if err := testCluster.updateMembers(testClient); err == nil {
		t.Fatalf("Removing member should fail")
	}
}

func TestUpdateMembersAddMember(t *testing.T) {
	t.Parallel()

	testContainersConfig := &container.Containers{
		DesiredState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	testContainers, err := testContainersConfig.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	testCluster := &cluster{
		containers: testContainers,
		members: map[string]Member{
			"foo": &member{
				config: &MemberConfig{
					Name:            "foo",
					PeerCertificate: "",
					PeerKey:         "",
					CACertificate:   utiltest.GenerateX509Certificate(t),
					Host: host.Host{
						DirectConfig: &direct.Config{},
					},
				},
			},
		},
	}

	testClient := &fakeClient{
		memberListF: func(context context.Context) (*clientv3.MemberListResponse, error) {
			return &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{},
			}, nil
		},
		memberAddF: func(context context.Context, peerURLs []string) (*clientv3.MemberAddResponse, error) {
			return nil, fmt.Errorf("expected")
		},
	}

	if err := testCluster.updateMembers(testClient); err == nil {
		t.Fatalf("Adding member should fail")
	}
}

// Deploy() tests.
func TestDeploy(t *testing.T) {
	t.Parallel()

	testContainersConfig := &container.Containers{
		DesiredState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	testContainers, err := testContainersConfig.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	testCluster := &cluster{
		containers: testContainers,
		members:    map[string]Member{},
	}

	err = testCluster.Deploy()
	if err == nil {
		t.Fatalf("Deploying bad containers should fail")
	}

	if !strings.Contains(err.Error(), "without knowing current state of the containers") {
		t.Fatalf("Deploying new cluster should not trigger updateMembers and fail on deploying, got: %v", err)
	}
}

func TestDeployUpdateMembers(t *testing.T) {
	t.Parallel()

	testContainersConfig := &container.Containers{
		PreviousState: container.ContainersState{
			"bar": getFakeHostConfiguredContainer(),
		},
		DesiredState: container.ContainersState{
			"foo": getFakeHostConfiguredContainer(),
		},
	}

	testContainers, err := testContainersConfig.New()
	if err != nil {
		t.Fatalf("Creating containers should succeed, got: %v", err)
	}

	testCluster := &cluster{
		containers: testContainers,
		members:    map[string]Member{},
	}

	err = testCluster.Deploy()
	if err == nil {
		t.Fatalf("Deploying should trigger updateMembers and fail")
	}

	expectedErrorMessage := "getting etcd client"
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Fatalf("Expected failure in client creation by error containing %q, got: %v", expectedErrorMessage, err)
	}
}

func TestClusterNewPKIIntegration(t *testing.T) {
	t.Parallel()

	pki := &pki.PKI{
		Etcd: &pki.Etcd{
			Peers: map[string]string{
				"test": "127.0.0.1",
			},
		},
	}

	if err := pki.Generate(); err != nil {
		t.Fatalf("Generating PKI should succeed, got: %v", err)
	}

	testClusterConfig := &Cluster{
		PKI: pki,
		Members: map[string]MemberConfig{
			"test": {
				PeerAddress: "127.0.0.1",
			},
		},
	}

	if _, err := testClusterConfig.New(); err != nil {
		t.Fatalf("Creating new cluster with valid PKI should succeed, got: %v", err)
	}
}
