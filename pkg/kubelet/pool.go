package kubelet

import (
	"fmt"
	"strconv"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Pool represents group of kubelet instances and their configuration.
type Pool struct {
	// User-configurable fields.
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
	CgroupDriver               string            `json:"cgroupDriver" yaml:"cgroupDriver"`
	NetworkPlugin              string            `json:"networkPlugin" yaml:"networkPlugin"`

	// Serializable fields.
	State container.ContainersState `json:"state" yaml:"state"`
}

// pool is a validated version of Pool.
type pool struct {
	containers container.Containers
}

// propagateKubelet fills given kubelet with values from Pool object.
func (p *Pool) propagateKubelet(k *Kubelet) {
	k.Image = util.PickString(k.Image, p.Image)
	k.BootstrapKubeconfig = util.PickString(k.BootstrapKubeconfig, p.BootstrapKubeconfig)
	k.KubernetesCACertificate = util.PickString(k.KubernetesCACertificate, p.KubernetesCACertificate)
	k.ClusterDNSIPs = util.PickStringSlice(k.ClusterDNSIPs, p.ClusterDNSIPs)
	k.Labels = util.PickStringMap(k.Labels, p.Labels)
	k.PrivilegedLabels = util.PickStringMap(k.PrivilegedLabels, p.PrivilegedLabels)
	k.PrivilegedLabelsKubeconfig = util.PickString(k.PrivilegedLabelsKubeconfig, p.PrivilegedLabelsKubeconfig)
	k.Taints = util.PickStringMap(k.Taints, p.Taints)
	k.CgroupDriver = util.PickString(k.CgroupDriver, p.CgroupDriver)
	k.NetworkPlugin = util.PickString(k.NetworkPlugin, p.NetworkPlugin)

	k.Host = host.BuildConfig(k.Host, host.Host{
		SSHConfig: p.SSH,
	})
}

// New validates kubelet pool configuration and fills all members with configured values.
func (p *Pool) New() (types.Resource, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate pool configuration: %w", err)
	}

	pool := &pool{
		containers: container.Containers{
			PreviousState: p.State,
			DesiredState:  make(container.ContainersState),
		},
	}

	for i := range p.Kubelets {
		k := &p.Kubelets[i]

		p.propagateKubelet(k)

		kubelet, _ := k.New()
		kubeletHcc, _ := kubelet.ToHostConfiguredContainer()

		pool.containers.DesiredState[strconv.Itoa(i)] = kubeletHcc
	}

	return pool, nil
}

// Validate validates Pool configuration.
//
// TODO add actual validation
func (p *Pool) Validate() error {
	for i := range p.Kubelets {
		// Make a copy of Kubelet struct to avoid modifying original one.
		k := p.Kubelets[i]

		p.propagateKubelet(&k)

		kubelet, err := k.New()
		if err != nil {
			return fmt.Errorf("failed to create kubelet object: %w", err)
		}

		if _, err := kubelet.ToHostConfiguredContainer(); err != nil {
			return fmt.Errorf("failed to generate kubelet container configuration: %w", err)
		}
	}

	return nil
}

// FromYaml allows to restore cluster state from YAML.
func FromYaml(c []byte) (types.Resource, error) {
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

// CheckCurrentState refreshes state of configured instances.
func (p *pool) CheckCurrentState() error {
	return p.containers.CheckCurrentState()
}

// Deploy checks current status of the pool and deploy configuration changes.
func (p *pool) Deploy() error {
	return p.containers.Deploy()
}
