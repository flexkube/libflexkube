// Package etcd allows to create and manage etcd clusters.
package etcd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.etcd.io/etcd/clientv3"
	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
	"github.com/flexkube/libflexkube/pkg/pki"
	"github.com/flexkube/libflexkube/pkg/types"
)

// defaultDialTimeout is default timeout value for etcd client.
const defaultDialTimeout = 5 * time.Second

// Cluster represents etcd cluster configuration and state from the user.
type Cluster struct {
	// User-configurable fields.
	Image         string            `json:"image,omitempty"`
	SSH           *ssh.Config       `json:"ssh,omitempty"`
	CACertificate types.Certificate `json:"caCertificate,omitempty"`
	Members       map[string]Member `json:"members,omitempty"`
	PKI           *pki.PKI          `json:"pki,omitempty"`

	// Serializable fields.
	State container.ContainersState `json:"state,omitempty"`
}

// cluster is executable version of Cluster, with validated fields and calculated containers.
type cluster struct {
	containers container.ContainersInterface
	members    map[string]*member
}

// propagateMember fills given Member's empty fields with fields from Cluster.
func (c *Cluster) propagateMember(i string, m *Member) {
	initialClusterArr := []string{}
	peerCertAllowedCNArr := []string{}

	for n, m := range c.Members {
		// If member has no name defined explicitly, use key passed as argument.
		name := util.PickString(m.Name, n)

		initialClusterArr = append(initialClusterArr, fmt.Sprintf("%s=https://%s:2380", name, m.PeerAddress))
		peerCertAllowedCNArr = append(peerCertAllowedCNArr, name)
	}

	sort.Strings(initialClusterArr)
	sort.Strings(peerCertAllowedCNArr)

	m.Name = util.PickString(m.Name, i)
	m.Image = util.PickString(m.Image, c.Image, defaults.EtcdImage)
	m.InitialCluster = util.PickString(m.InitialCluster, strings.Join(initialClusterArr, ","))
	m.PeerCertAllowedCN = util.PickString(m.PeerCertAllowedCN, strings.Join(peerCertAllowedCNArr, ","))
	m.CACertificate = m.CACertificate.Pick(c.CACertificate)

	// PKI integration.
	if c.PKI != nil && c.PKI.Etcd != nil {
		e := c.PKI.Etcd

		m.CACertificate = m.CACertificate.Pick(c.CACertificate, e.CA.X509Certificate)

		if c, ok := e.PeerCertificates[m.Name]; ok {
			m.PeerCertificate = m.PeerCertificate.Pick(c.X509Certificate)
			m.PeerKey = m.PeerKey.Pick(c.PrivateKey)
		}

		if c, ok := e.ServerCertificates[m.Name]; ok {
			m.ServerCertificate = m.ServerCertificate.Pick(c.X509Certificate)
			m.ServerKey = m.ServerKey.Pick(c.PrivateKey)
		}
	}

	m.ServerAddress = util.PickString(m.ServerAddress, m.PeerAddress)

	m.Host = host.BuildConfig(m.Host, host.Host{
		SSHConfig: c.SSH,
	})

	if len(c.State) == 0 {
		m.NewCluster = true
	}
}

// New validates etcd cluster configuration and fills members with default and computed values.
func (c *Cluster) New() (types.Resource, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate cluster configuration: %w", err)
	}

	cc := container.Containers{
		PreviousState: c.State,
		DesiredState:  make(container.ContainersState),
	}

	cluster := &cluster{
		members: map[string]*member{},
	}

	for n, m := range c.Members {
		m := m
		c.propagateMember(n, &m)

		mem, _ := m.New()
		hcc, _ := mem.ToHostConfiguredContainer()

		cc.DesiredState[n] = hcc

		cluster.members[n] = mem.(*member)
	}

	co, _ := cc.New()

	cluster.containers = co

	return cluster, nil
}

// Validate validates Cluster configuration.
func (c *Cluster) Validate() error {
	if len(c.Members) == 0 && c.State == nil {
		return fmt.Errorf("either members or previous state needs to be defined")
	}

	var errors util.ValidateError

	cc := container.Containers{
		PreviousState: c.State,
		DesiredState:  make(container.ContainersState),
	}

	for n, m := range c.Members {
		m := m
		c.propagateMember(n, &m)

		mem, err := m.New()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to validate member '%s': %w", n, err))

			continue
		}

		hcc, err := mem.ToHostConfiguredContainer()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to validate member '%s' container: %w", n, err))

			continue
		}

		cc.DesiredState[n] = hcc
	}

	if _, err := cc.New(); err != nil {
		errors = append(errors, fmt.Errorf("failed validating containers object: %w", err))
	}

	return errors.Return()
}

// FromYaml allows to restore cluster state from YAML.
func FromYaml(c []byte) (types.Resource, error) {
	return types.ResourceFromYaml(c, &Cluster{})
}

// StateToYaml allows to dump cluster state to YAML, so it can be restored later.
func (c *cluster) StateToYaml() ([]byte, error) {
	return yaml.Marshal(Cluster{State: c.containers.ToExported().PreviousState})
}

// CheckCurrentState refreshes current state of the cluster.
func (c *cluster) CheckCurrentState() error {
	if err := c.containers.CheckCurrentState(); err != nil {
		return fmt.Errorf("failed checking current state of etcd cluster: %w", err)
	}

	return nil
}

// getExistingEndpoints returns list of already deployed etcd endpoints.
func (c *cluster) getExistingEndpoints() []string {
	endpoints := []string{}

	for i, m := range c.members {
		if _, ok := c.containers.ToExported().PreviousState[i]; !ok {
			continue
		}

		endpoints = append(endpoints, fmt.Sprintf("%s:2379", m.peerAddress))
	}

	return endpoints
}

func (c *cluster) firstMember() (*member, error) {
	var m *member

	for i := range c.members {
		m = c.members[i]
		break
	}

	if m == nil {
		return nil, fmt.Errorf("no members defined")
	}

	return m, nil
}

func (c *cluster) getClient() (etcdClient, error) {
	m, err := c.firstMember()
	if err != nil {
		return nil, fmt.Errorf("failed getting member object: %w", err)
	}

	endpoints, err := m.forwardEndpoints(c.getExistingEndpoints())
	if err != nil {
		return nil, fmt.Errorf("failed forwarding endpoints: %w", err)
	}

	return m.getEtcdClient(endpoints)
}

type etcdClient interface {
	MemberList(context context.Context) (*clientv3.MemberListResponse, error)
	MemberAdd(context context.Context, peerURLs []string) (*clientv3.MemberAddResponse, error)
	MemberRemove(context context.Context, id uint64) (*clientv3.MemberRemoveResponse, error)
	Close() error
}

func (c *cluster) membersToRemove() []string {
	m := []string{}

	e := c.containers.ToExported()

	for i := range e.PreviousState {
		if _, ok := e.DesiredState[i]; !ok {
			m = append(m, i)
		}
	}

	return m
}

func (c *cluster) membersToAdd() []string {
	m := []string{}

	e := c.containers.ToExported()

	for i := range e.DesiredState {
		if _, ok := e.PreviousState[i]; !ok {
			m = append(m, i)
		}
	}

	return m
}

// updateMembers adds and remove members from the cluster according to the configuration.
func (c *cluster) updateMembers(cli etcdClient) error {
	for _, name := range c.membersToRemove() {
		m := &member{
			name: name,
		}

		if err := m.remove(cli); err != nil {
			return fmt.Errorf("failed removing member: %w", err)
		}
	}

	for _, m := range c.membersToAdd() {
		if err := c.members[m].add(cli); err != nil {
			return fmt.Errorf("failed adding member: %w", err)
		}
	}

	return nil
}

// Deploy refreshes current state of the cluster and deploys detected changes.
func (c *cluster) Deploy() error {
	e := c.containers.ToExported()

	// If we create new cluster or destroy entire cluster, just start deploying.
	if len(e.PreviousState) != 0 && len(e.DesiredState) != 0 {
		// Build client, so we can pass it around.
		cli, err := c.getClient()
		if err != nil {
			return fmt.Errorf("failed getting etcd client: %w", err)
		}

		if err := c.updateMembers(cli); err != nil {
			return fmt.Errorf("failed to update members before deploying: %w", err)
		}

		if err := cli.Close(); err != nil {
			return fmt.Errorf("failed to close etcd client: %w", err)
		}
	}

	return c.containers.Deploy()
}

// Containers implement types.Resource interface.
func (c *cluster) Containers() container.ContainersInterface {
	return c.containers
}
