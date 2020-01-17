package etcd

import (
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Cluster represents etcd cluster configuration and state from the user.
type Cluster struct {
	// User-configurable fields.
	Image         string            `json:"image" yaml:"image"`
	SSH           *ssh.Config       `json:"ssh" yaml:"ssh"`
	CACertificate string            `json:"caCertificate" yaml:"caCertificate"`
	Members       map[string]Member `json:"members" yaml:"members"`

	// Serializable fields.
	State container.ContainersState `json:"state" yaml:"state"`
}

// cluster is executable version of Cluster, with validated fields and calculated containers.
type cluster struct {
	image         string
	ssh           *ssh.Config
	caCertificate string
	containers    container.Containers
}

// New validates etcd cluster configuration and fills members with default and computed values.
func (c *Cluster) New() (types.Resource, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate cluster configuration: %w", err)
	}

	cluster := &cluster{
		image:         c.Image,
		ssh:           c.SSH,
		caCertificate: c.CACertificate,
		containers: container.Containers{
			PreviousState: c.State,
			DesiredState:  make(container.ContainersState),
		},
	}

	initialClusterArr := []string{}
	peerCertAllowedCNArr := []string{}

	for n, m := range c.Members {
		initialClusterArr = append(initialClusterArr, fmt.Sprintf("%s=https://%s:2380", fmt.Sprintf("etcd-%s", n), m.PeerAddress))
		peerCertAllowedCNArr = append(peerCertAllowedCNArr, fmt.Sprintf("etcd-%s", n))
	}

	initialCluster := strings.Join(initialClusterArr, ",")
	peerCertAllowedCN := strings.Join(peerCertAllowedCNArr, ",")

	for n, m := range c.Members {
		if m.Name == "" {
			m.Name = fmt.Sprintf("etcd-%s", n)
		}

		if m.Image == "" && c.Image != "" {
			m.Image = c.Image
		}

		if m.InitialCluster == "" {
			m.InitialCluster = initialCluster
		}

		if m.PeerCertAllowedCN == "" {
			m.PeerCertAllowedCN = peerCertAllowedCN
		}

		// TODO find better way to handle defaults!!!
		if m.Host == nil || (m.Host.DirectConfig == nil && m.Host.SSHConfig == nil) {
			m.Host = &host.Host{
				DirectConfig: &direct.Config{},
			}
		}

		m.Host.SSHConfig = ssh.BuildConfig(m.Host.SSHConfig, c.SSH)

		if m.CACertificate == "" && c.CACertificate != "" {
			m.CACertificate = c.CACertificate
		}

		member, err := m.New()
		if err != nil {
			return nil, fmt.Errorf("that was unexpected: %w", err)
		}

		hcc, err := member.ToHostConfiguredContainer()
		if err != nil {
			return nil, fmt.Errorf("failed building container configuration: %w", err)
		}

		cluster.containers.DesiredState[n] = hcc
	}

	return cluster, nil
}

// Validate validates Cluster configuration.
func (c *Cluster) Validate() error {
	if len(c.Members) == 0 && c.State == nil {
		return fmt.Errorf("either members or previous state needs to be defined")
	}

	for n, m := range c.Members {
		// TODO validate only fills default fields here which will be done in a separated step anyway.
		// we should make this a function!
		if m.Name == "" {
			m.Name = n
		}

		if m.Host == nil || m.Host.DirectConfig == nil || m.Host.SSHConfig == nil {
			m.Host = &host.Host{
				DirectConfig: &direct.Config{},
			}
		}

		if err := m.Validate(); err != nil {
			return fmt.Errorf("failed to validate member '%s': %w", n, err)
		}
	}

	return nil
}

// FromYaml allows to restore cluster state from YAML.
func FromYaml(c []byte) (types.Resource, error) {
	cluster := &Cluster{}
	if err := yaml.Unmarshal(c, &cluster); err != nil {
		return nil, fmt.Errorf("failed to parse input yaml: %w", err)
	}

	cl, err := cluster.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster object: %w", err)
	}

	return cl, nil
}

// StateToYaml allows to dump cluster state to YAML, so it can be restored later.
func (c *cluster) StateToYaml() ([]byte, error) {
	return yaml.Marshal(Cluster{State: c.containers.PreviousState})
}

// CheckCurrentState refreshes current state of the cluster.
func (c *cluster) CheckCurrentState() error {
	return c.containers.CheckCurrentState()
}

// Deploy refreshes current state of the cluster and deploys detected changes.
func (c *cluster) Deploy() error {
	return c.containers.Deploy()
}
