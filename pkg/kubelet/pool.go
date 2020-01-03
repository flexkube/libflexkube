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
	Image                      string            `json:"image" yaml:"image"`
	SSH                        *ssh.Config       `json:"ssh" yaml:"ssh"`
	BootstrapKubeconfig        string            `json:"bootstrapKubeconfig" yaml:"bootstrapKubeconfig"`
	Kubelets                   []Kubelet         `json:"kubelets" yaml:"kubelets"`
	KubernetesCACertificate    string            `json:"kubernetesCACertificate" yaml:"kubernetesCACertificate"`
	ClusterDNSIPs              []string          `json:"clusterDNSIPs" yaml:"clusterDNSIPs"`
	Taints                     map[string]string `json:"taints" yaml:"taints"`
	Labels                     map[string]string `json:"labels" yaml:"labels"`
	PrivilegedLabels           map[string]string `json:"privilegedLabels" yaml:"privilegedLabels"`
	PrivilegedLabelsKubeconfig string            `json:"privilegedLabelsKubeconfig" yaml:"privilegedLabelsKubeconfig"`

	// Serializable fields
	State container.ContainersState `json:"state" yaml:"state"`
}

// pool is a validated version of Pool
type pool struct {
	image      string
	ssh        *ssh.Config
	containers container.Containers
}

// New validates kubelet pool configuration and fills all members with configured values
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

		if len(k.ClusterDNSIPs) == 0 && len(p.ClusterDNSIPs) > 0 {
			k.ClusterDNSIPs = p.ClusterDNSIPs
		}

		if len(k.Labels) == 0 && len(p.Labels) > 0 {
			k.Labels = p.Labels
		}

		if len(k.PrivilegedLabels) == 0 && len(p.PrivilegedLabels) > 0 {
			k.PrivilegedLabels = p.PrivilegedLabels
		}

		if k.PrivilegedLabelsKubeconfig == "" && p.PrivilegedLabelsKubeconfig != "" {
			k.PrivilegedLabelsKubeconfig = p.PrivilegedLabelsKubeconfig
		}

		if len(k.Taints) == 0 && len(p.Taints) > 0 {
			k.Taints = p.Taints
		}

		// TODO find better way to handle defaults!!!
		if k.Host.DirectConfig == nil && k.Host.SSHConfig == nil && p.SSH == nil {
			k.Host = host.Host{
				DirectConfig: &direct.Config{},
			}
		}

		if p.SSH != nil {
			k.Host.SSHConfig = ssh.BuildConfig(k.Host.SSHConfig, p.SSH)
		}

		kubelet, err := k.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create kubelet object: %w", err)
		}

		pool.containers.DesiredState[strconv.Itoa(i)] = kubelet.ToHostConfiguredContainer()
	}

	return pool, nil
}

// Validate validates Pool configuration
//
// TODO add actual validation
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
		return nil, fmt.Errorf("failed to create pool object: %w", err)
	}

	return p, nil
}

// StateToYaml allows to dump cluster state to YAML, so it can be restored later.
func (p *pool) StateToYaml() ([]byte, error) {
	return yaml.Marshal(Pool{State: p.containers.PreviousState})
}

// CheckCurrentState refreshes state of configured instances
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

// Deploy checks current status of the pool and deploy configuration changes
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
