package etcd

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/invidian/libflexkube/pkg/container"
	"github.com/invidian/libflexkube/pkg/host"
	"github.com/invidian/libflexkube/pkg/host/transport/direct"
	"github.com/invidian/libflexkube/pkg/host/transport/ssh"
)

// Cluster represents etcd cluster configuration and state from the user
type Cluster struct {
	// User-configurable fields
	Image             string            `json:"image,omitempty" yaml:"image,omitempty"`
	SSH               *ssh.SSHConfig    `json:"ssh,omitempty" yaml:"ssh,omitempty"`
	PeerCACertificate string            `json:"peerCACertificate,omitempty" yaml:"peerCACertificate,omitempty"`
	Members           map[string]Member `json:"members,omitempty" yaml:"members,omitempty"`

	// Serializable fields
	State container.ContainersState `json:"state:omitempty" yaml:"state,omitempty"`
}

// cluster is executable version of Cluster, with validated fields and calculated containers
type cluster struct {
	image             string
	ssh               *ssh.SSHConfig
	peerCACertificate string
	containers        container.Containers
}

func (c *Cluster) New() (*cluster, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate cluster configuration: %w", err)
	}

	cluster := &cluster{
		image:             c.Image,
		ssh:               c.SSH,
		peerCACertificate: c.PeerCACertificate,
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
				DirectConfig: &direct.DirectConfig{},
			}
		}
		if m.Host != nil && m.Host.SSHConfig != nil && m.Host.SSHConfig.PrivateKey == "" && c.SSH != nil && c.SSH.PrivateKey != "" {
			m.Host.SSHConfig.PrivateKey = c.SSH.PrivateKey
		}

		if m.Host != nil && m.Host.SSHConfig != nil && m.Host.SSHConfig.User == "" && c.SSH != nil && c.SSH.User != "" {
			m.Host.SSHConfig.User = c.SSH.User
		}
		if m.Host != nil && m.Host.SSHConfig != nil && m.Host.SSHConfig.User == "" {
			m.Host.SSHConfig.User = "root"
		}

		if m.Host != nil && m.Host.SSHConfig != nil && m.Host.SSHConfig.ConnectionTimeout == "" && c.SSH != nil && c.SSH.ConnectionTimeout != "" {
			m.Host.SSHConfig.ConnectionTimeout = c.SSH.ConnectionTimeout
		}
		if m.Host != nil && m.Host.SSHConfig != nil && m.Host.SSHConfig.ConnectionTimeout == "" {
			m.Host.SSHConfig.ConnectionTimeout = "30s"
		}

		if m.Host != nil && m.Host.SSHConfig != nil && m.Host.SSHConfig.Port == 0 && c.SSH != nil && c.SSH.Port != 0 {
			m.Host.SSHConfig.Port = c.SSH.Port
		}
		if m.Host != nil && m.Host.SSHConfig != nil && m.Host.SSHConfig.Port == 0 {
			m.Host.SSHConfig.Port = 22
		}

		if m.PeerCACertificate == "" && c.PeerCACertificate != "" {
			m.PeerCACertificate = c.PeerCACertificate
		}

		member, err := m.New()
		if err != nil {
			return nil, fmt.Errorf("that was unexpected: %w", err)
		}
		cluster.containers.DesiredState[n] = member.ToHostConfiguredContainer()
	}

	return cluster, nil
}

func (c *Cluster) Validate() error {
	if len(c.Members) < 0 && c.State == nil {
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
				DirectConfig: &direct.DirectConfig{},
			}
		}
		if err := m.Validate(); err != nil {
			return fmt.Errorf("failed to validate member '%s': %w", n, err)
		}
	}

	return nil
}

// FromYaml allows to restore cluster state from YAML.
func FromYaml(c []byte) (*cluster, error) {
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

func (c *cluster) CheckCurrentState() error {
	containers, err := c.containers.New()
	if err != nil {
		return err
	}
	if err := containers.CheckCurrentState(); err != nil {
		return err
	}
	c.containers = *containers.ToExported()
	return nil
}

func (c *cluster) Deploy() error {
	containers, err := c.containers.New()
	if err != nil {
		return err
	}
	// TODO Deploy shouldn't refresh the state. However, due to how we handle exported/unexported
	// structs to enforce validation of objects, we lose current state, as we want it to be computed.
	// On the other hand, maybe it's a good thing to call it once we execute. This way we could compare
	// the plan user agreed to execute with plan calculated right before the execution and fail early if they
	// differ.
	// This is similar to what terraform is doing and may cause planning to run several times, so it may require
	// some optimization.
	// Alternatively we can have serializable plan and a knob in execute command to control wheather we should
	// make additional validation or not.
	if err := containers.CheckCurrentState(); err != nil {
		return err
	}
	if err := containers.Execute(); err != nil {
		return err
	}
	c.containers = *containers.ToExported()
	return nil
}
