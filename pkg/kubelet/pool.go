package kubelet

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

// Pool represents group of kubelet instances and their configuration
type Pool struct {
	// User-configurable fields
	Image                   string         `json:"image,omitempty" yaml:"image,omitempty"`
	SSH                     *ssh.SSHConfig `json:"ssh,omitempty" yaml:"ssh,omitempty"`
	BootstrapKubeconfig     string         `json:"bootstrapKubeconfig,omitempty" yaml:"bootstrapKubeconfig,omitempty"`
	Kubelets                []Kubelet      `json:"kubelets,omitempty" yaml:"kubelets,omitempty"`
	KubernetesCACertificate string         `json:"kubernetesCACertificate,omitempty" yaml:"kubernetesCACertificate,omitempty"`
	ClusterDNSIPs           []string       `json:"clusterDNSIPs,omitempty" yaml:"clusterDNSIPs,omitempty"`

	// Serializable fields
	State container.ContainersState `json:"state:omitempty" yaml:"state,omitempty"`
}

type pool struct {
	image      string
	ssh        *ssh.SSHConfig
	containers container.Containers
}

func (p *Pool) New() (*pool, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate pool configuration: %w", err)
	}

	pool := &pool{
		image: p.Image,
		ssh:   p.SSH,
		containers: container.Containers{
			PreviousState: p.State,
			DesiredState:  make(container.ContainersState),
		},
	}

	for i, k := range p.Kubelets {
		if k.Image == "" && p.Image != "" {
			k.Image = p.Image
		}
		if k.BootstrapKubeconfig == "" && p.BootstrapKubeconfig != "" {
			k.BootstrapKubeconfig = p.BootstrapKubeconfig
		}
		if k.KubernetesCACertificate == "" && p.KubernetesCACertificate != "" {
			k.KubernetesCACertificate = p.KubernetesCACertificate
		}
		if len(k.ClusterDNSIPs) <= 0 && len(p.ClusterDNSIPs) > 0 {
			k.ClusterDNSIPs = p.ClusterDNSIPs
		}

		// TODO find better way to handle defaults!!!
		if k.Host == nil || (k.Host.DirectConfig == nil && k.Host.SSHConfig == nil) {
			k.Host = &host.Host{
				DirectConfig: &direct.Config{},
			}
		}
		if k.Host != nil && k.Host.SSHConfig != nil && k.Host.SSHConfig.PrivateKey == "" && p.SSH != nil && p.SSH.PrivateKey != "" {
			k.Host.SSHConfig.PrivateKey = p.SSH.PrivateKey
		}

		if k.Host != nil && k.Host.SSHConfig != nil && k.Host.SSHConfig.User == "" && p.SSH != nil && p.SSH.User != "" {
			k.Host.SSHConfig.User = p.SSH.User
		}
		if k.Host != nil && k.Host.SSHConfig != nil && k.Host.SSHConfig.User == "" {
			k.Host.SSHConfig.User = "root"
		}

		if k.Host != nil && k.Host.SSHConfig != nil && k.Host.SSHConfig.ConnectionTimeout == "" && p.SSH != nil && p.SSH.ConnectionTimeout != "" {
			k.Host.SSHConfig.ConnectionTimeout = p.SSH.ConnectionTimeout
		}
		if k.Host != nil && k.Host.SSHConfig != nil && k.Host.SSHConfig.ConnectionTimeout == "" {
			k.Host.SSHConfig.ConnectionTimeout = "30s"
		}

		if k.Host != nil && k.Host.SSHConfig != nil && k.Host.SSHConfig.Port == 0 && p.SSH != nil && p.SSH.Port != 0 {
			k.Host.SSHConfig.Port = p.SSH.Port
		}
		if k.Host != nil && k.Host.SSHConfig != nil && k.Host.SSHConfig.Port == 0 {
			k.Host.SSHConfig.Port = 22
		}

		kubelet, err := k.New()
		if err != nil {
			return nil, fmt.Errorf("that was unexpected: %w", err)
		}
		pool.containers.DesiredState[strconv.Itoa(i)] = kubelet.ToHostConfiguredContainer()
	}

	return pool, nil
}

// TODO add validation
func (p *Pool) Validate() error {
	return nil
}

// FromYaml allows to restore cluster state from YAML.
func FromYaml(c []byte) (*pool, error) {
	pool := &Pool{}
	if err := yaml.Unmarshal(c, &pool); err != nil {
		return nil, fmt.Errorf("failed to parse input yaml: %w", err)
	}
	p, err := pool.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster object: %w", err)
	}
	return p, nil
}

// StateToYaml allows to dump cluster state to YAML, so it can be restored later.
func (p *pool) StateToYaml() ([]byte, error) {
	return yaml.Marshal(Pool{State: p.containers.PreviousState})
}

func (p *pool) CheckCurrentState() error {
	containers, err := p.containers.New()
	if err != nil {
		return err
	}
	if err := containers.CheckCurrentState(); err != nil {
		return err
	}
	p.containers = *containers.ToExported()
	return nil
}

func (p *pool) Deploy() error {
	containers, err := p.containers.New()
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
	// Alternatively we can have serializable plan and a knob in execute command to control whether we should
	// make additional validation or not.
	if err := containers.CheckCurrentState(); err != nil {
		return err
	}
	if err := containers.Execute(); err != nil {
		return err
	}
	p.containers = *containers.ToExported()
	return nil
}
