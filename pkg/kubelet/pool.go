// Package kubelet implements logic needed for creating and managing kubelet instances
// running as containers.
package kubelet

import (
	"fmt"
	"strconv"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Pool represents group of kubelet instances and their configuration.
type Pool struct {
	// User-configurable fields.
	Image                   string                 `json:"image,omitempty"`
	SSH                     *ssh.Config            `json:"ssh,omitempty"`
	BootstrapConfig         *client.Config         `json:"bootstrapConfig,omitempty"`
	Kubelets                []Kubelet              `json:"kubelets,omitempty"`
	KubernetesCACertificate types.Certificate      `json:"kubernetesCACertificate,omitempty"`
	ClusterDNSIPs           []string               `json:"clusterDNSIPs,omitempty"`
	Taints                  map[string]string      `json:"taints,omitempty"`
	Labels                  map[string]string      `json:"labels,omitempty"`
	PrivilegedLabels        map[string]string      `json:"privilegedLabels,omitempty"`
	AdminConfig             *client.Config         `json:"adminConfig,omitempty"`
	CgroupDriver            string                 `json:"cgroupDriver,omitempty"`
	NetworkPlugin           string                 `json:"networkPlugin,omitempty"`
	SystemReserved          map[string]string      `json:"systemReserved,omitempty"`
	KubeReserved            map[string]string      `json:"kubeReserved,omitempty"`
	HairpinMode             string                 `json:"hairpinMode,omitempty"`
	VolumePluginDir         string                 `json:"volumePluginDir,omitempty"`
	ExtraMounts             []containertypes.Mount `json:"extraMounts,omitempty"`

	// Serializable fields.
	State container.ContainersState `json:"state,omitempty"`
}

// pool is a validated version of Pool.
type pool struct {
	containers container.ContainersInterface
}

// propagateKubelet fills given kubelet with values from Pool object.
func (p *Pool) propagateKubelet(k *Kubelet) {
	k.Image = util.PickString(k.Image, p.Image)
	k.KubernetesCACertificate = types.Certificate(util.PickString(string(k.KubernetesCACertificate), string(p.KubernetesCACertificate)))
	k.ClusterDNSIPs = util.PickStringSlice(k.ClusterDNSIPs, p.ClusterDNSIPs)
	k.Labels = util.PickStringMap(k.Labels, p.Labels)
	k.PrivilegedLabels = util.PickStringMap(k.PrivilegedLabels, p.PrivilegedLabels)
	k.Taints = util.PickStringMap(k.Taints, p.Taints)
	k.CgroupDriver = util.PickString(k.CgroupDriver, p.CgroupDriver)
	k.NetworkPlugin = util.PickString(k.NetworkPlugin, p.NetworkPlugin)
	k.SystemReserved = util.PickStringMap(k.SystemReserved, p.SystemReserved)
	k.KubeReserved = util.PickStringMap(k.KubeReserved, p.KubeReserved)
	k.HairpinMode = util.PickString(k.HairpinMode, p.HairpinMode)
	k.VolumePluginDir = util.PickString(k.VolumePluginDir, p.VolumePluginDir)

	if len(k.ExtraMounts) == 0 {
		k.ExtraMounts = p.ExtraMounts
	}

	k.Host = host.BuildConfig(k.Host, host.Host{
		SSHConfig: p.SSH,
	})

	if k.BootstrapConfig == nil && p.BootstrapConfig != nil {
		k.BootstrapConfig = p.BootstrapConfig
	}

	if k.AdminConfig == nil && p.AdminConfig != nil {
		k.AdminConfig = p.AdminConfig
	}

	if k.BootstrapConfig != nil && k.BootstrapConfig.CACertificate == "" && p.KubernetesCACertificate != "" {
		k.BootstrapConfig.CACertificate = p.KubernetesCACertificate
	}

	if k.AdminConfig != nil && k.AdminConfig.CACertificate == "" && p.KubernetesCACertificate != "" {
		k.AdminConfig.CACertificate = p.KubernetesCACertificate
	}
}

// New validates kubelet pool configuration and fills all members with configured values.
func (p *Pool) New() (types.Resource, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate pool configuration: %w", err)
	}

	cc := &container.Containers{
		PreviousState: p.State,
		DesiredState:  make(container.ContainersState),
	}

	for i := range p.Kubelets {
		k := &p.Kubelets[i]

		p.propagateKubelet(k)

		kubelet, _ := k.New()
		kubeletHcc, _ := kubelet.ToHostConfiguredContainer()

		cc.DesiredState[strconv.Itoa(i)] = kubeletHcc
	}

	c, _ := cc.New()

	return &pool{
		containers: c,
	}, nil
}

// Validate validates Pool configuration.
//
// TODO: Add actual validation.
func (p *Pool) Validate() error {
	cc := &container.Containers{
		PreviousState: p.State,
		DesiredState:  make(container.ContainersState),
	}

	for i := range p.Kubelets {
		// Make a copy of Kubelet struct to avoid modifying original one.
		k := p.Kubelets[i]

		p.propagateKubelet(&k)

		kubelet, err := k.New()
		if err != nil {
			return fmt.Errorf("failed to create kubelet object: %w", err)
		}

		hcc, err := kubelet.ToHostConfiguredContainer()
		if err != nil {
			return fmt.Errorf("failed to generate kubelet container configuration: %w", err)
		}

		cc.DesiredState[strconv.Itoa(i)] = hcc
	}

	if _, err := cc.New(); err != nil {
		return fmt.Errorf("failed validating containers configuration: %w", err)
	}

	return nil
}

// FromYaml allows to restore cluster state from YAML.
func FromYaml(c []byte) (types.Resource, error) {
	return types.ResourceFromYaml(c, &Pool{})
}

// StateToYaml allows to dump cluster state to YAML, so it can be restored later.
func (p *pool) StateToYaml() ([]byte, error) {
	return yaml.Marshal(Pool{State: p.containers.ToExported().PreviousState})
}

// CheckCurrentState refreshes state of configured instances.
func (p *pool) CheckCurrentState() error {
	return p.containers.CheckCurrentState()
}

// Deploy checks current status of the pool and deploy configuration changes.
func (p *pool) Deploy() error {
	return p.containers.Deploy()
}

// Containers implement types.Resource interface.
func (p *pool) Containers() container.ContainersInterface {
	return p.containers
}
