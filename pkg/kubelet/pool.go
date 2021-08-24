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
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/pki"
	"github.com/flexkube/libflexkube/pkg/types"
)

const (
	// DefaultNetworkPlugin is a default NetworkPlugin configured for kubelets.
	DefaultNetworkPlugin = "cni"
	// DefaultHairpinMode is a default HairpinMode configured for kubelets.
	DefaultHairpinMode = "hairpin-veth"
)

// Pool represents group of kubelet instances and their configuration.
//
// It implements types.ResourceConfig interface and via types.Resource interface which
// allows to manage full lifecycle of all kubelet processes on the cluster.
//
// It handles updating kubelet version, updating configuration and used flags. It also
// allows to set "privileged" Node labels like 'node-role.kubernetes.io/master', which
// kubelet cannot set by itself.
type Pool struct {
	// Image allows to set Docker image with tag, which will be used by all kubelets,
	// if they have no image set. If empty, hyperkube image defined in pkg/defaults
	// will be used.
	//
	// Example value: 'k8s.gcr.io/hyperkube:v1.18.3'.
	//
	// This field is optional.
	Image string `json:"image,omitempty"`

	// SSH stores common SSH configuration for all kubelets and will be merged with kubelets
	// SSH configuration. If kubelet has some SSH fields defined, they take precedence over
	// this block.
	//
	// If you use same username or port for all members, it is recommended to have it defined
	// here to avoid repetition in the configuration.
	//
	// This field is optional.
	SSH *ssh.Config `json:"ssh,omitempty"`

	// BootstrapConfig contains kubelet bootstrap kubeconfig configuration, including
	// bootstrap token and Kubernetes API server address.
	//
	// This field is optional, if each kubelet instance has this field set.
	BootstrapConfig *client.Config `json:"bootstrapConfig,omitempty"`

	// Kubelets holds a list of kubelet instances to create.
	Kubelets []Kubelet `json:"kubelets,omitempty"`

	// KubernetesCACertificate holds Kubernetes X.509 CA certificate, PEM encoded, which will
	// be used by kubelets to verify Kubernetes API server they talk to.
	KubernetesCACertificate types.Certificate `json:"kubernetesCACertificate,omitempty"`

	// ClusterDNSIPs is a list of IP addresses, which will be used in pods for as DNS servers
	// to allow cluster names resolution. This is usually set to 10th address of service CIDR,
	// so if your service CIDR is 11.0.0.0/16, it should be 11.0.0.10.
	//
	// Example value: '11.0.0.10'.
	ClusterDNSIPs []string `json:"clusterDNSIPs,omitempty"`

	// Taints is a list of taints, which should be set for all kubelets.
	Taints map[string]string `json:"taints,omitempty"`

	// Labels is a list of labels, which should be used when kubelet registers Node object into
	// cluster.
	Labels map[string]string `json:"labels,omitempty"`

	// PrivilegedLabels is a list of labels, which kubelet cannot apply by itself due to node
	// isolation restrictions, but administrator wants to set them. One of such labels is
	// 'node-role.kubernetes.io/master', which gives node a master role, which attract pods
	// which has access to cluster secrets, like kube-apiserver etc.
	PrivilegedLabels map[string]string `json:"privilegedLabels,omitempty"`

	// AdminConfig is a simplified version of kubeconfig, which will be used for applying
	// privileged labels while the pool is created/updated.
	AdminConfig *client.Config `json:"adminConfig,omitempty"`

	// CgroupDriver configures cgroup driver to be used by the kubelet. It must be the same
	// as configured for container runtime used by the kubelet.
	CgroupDriver string `json:"cgroupDriver,omitempty"`

	// NetworkPlugin defines which network solution should be used by kubelet to assign
	// IP addresses to the pods. By default, 'cni' is used. Also 'kubelet' is a valid value.
	NetworkPlugin string `json:"networkPlugin,omitempty"`

	// SystemReserved configures, how much resources kubelet should mark as used by the operating
	// system.
	SystemReserved map[string]string `json:"systemReserved,omitempty"`

	// KubeReserved configures, how much resources kubelet should mark as used by the Kubernetes
	// itself on the node.
	KubeReserved map[string]string `json:"kubeReserved,omitempty"`

	// HairpinMode controls kubelet hairpin mode.
	HairpinMode string `json:"hairpinMode,omitempty"`

	// VolumePluginDir configures, where Flexvolume plugins should be installed. It will be used
	// unless kubelet instance define it's own VolumePluginDir.
	VolumePluginDir string `json:"volumePluginDir,omitempty"`

	// ExtraMounts defines extra mounts from host filesystem, which should be added to kubelet
	// containers. It will be used unless kubelet instance define it's own extra mounts.
	ExtraMounts []containertypes.Mount `json:"extraMounts,omitempty"`

	// PKI field allows to use PKI resource for managing all kubernetes certificates. It will be
	// used for kubelets configuration, if they don't have certificates defined.
	PKI *pki.PKI `json:"pki,omitempty"`

	// Serializable fields.
	State container.ContainersState `json:"state,omitempty"`

	// WaitForNodeReady controls, if deploy should wait until node becomes ready.
	WaitForNodeReady bool `json:"waitForNodeReady,omitempty"`

	// ExtraArgs defines additional flags which will be added to the kubelet process.
	ExtraArgs []string `json:"extraArgs,omitempty"`
}

// pool is a validated version of Pool.
type pool struct {
	containers container.ContainersInterface
}

// pkiIntegration merges certificates from PKI into pool configuration.
func (p *Pool) pkiIntegration() {
	if p.PKI == nil || p.PKI.Kubernetes == nil {
		return
	}

	if p.PKI.Kubernetes.CA != nil && p.KubernetesCACertificate == "" {
		p.KubernetesCACertificate = p.PKI.Kubernetes.CA.X509Certificate
	}

	if p.AdminConfig == nil {
		return
	}

	if p.AdminConfig.ClientCertificate == "" && p.PKI.Kubernetes.AdminCertificate != nil {
		p.AdminConfig.ClientCertificate = p.PKI.Kubernetes.AdminCertificate.X509Certificate
	}

	if p.AdminConfig.ClientKey == "" && p.PKI.Kubernetes.AdminCertificate != nil {
		p.AdminConfig.ClientKey = p.PKI.Kubernetes.AdminCertificate.PrivateKey
	}
}

// kubeletPKIIntegration merges certificates from PKI into given kubelet configuration.
func (p *Pool) kubeletPKIIntegration(k *Kubelet) {
	k.KubernetesCACertificate = types.Certificate(util.PickString(string(k.KubernetesCACertificate), string(p.KubernetesCACertificate)))

	if p.BootstrapConfig != nil && k.BootstrapConfig == nil {
		k.BootstrapConfig = p.BootstrapConfig
	}

	if p.AdminConfig != nil && k.AdminConfig == nil {
		k.AdminConfig = p.AdminConfig
	}

	if k.BootstrapConfig != nil && k.BootstrapConfig.CACertificate == "" {
		k.BootstrapConfig.CACertificate = p.KubernetesCACertificate
	}

	if k.AdminConfig != nil && k.AdminConfig.CACertificate == "" {
		k.AdminConfig.CACertificate = p.KubernetesCACertificate
	}
}

// propagateKubelet fills given kubelet with values from Pool object.
func (p *Pool) propagateKubelet(k *Kubelet) {
	k.Image = util.PickString(k.Image, p.Image)
	k.ClusterDNSIPs = util.PickStringSlice(k.ClusterDNSIPs, p.ClusterDNSIPs)
	k.Labels = util.PickStringMap(k.Labels, p.Labels)
	k.PrivilegedLabels = util.PickStringMap(k.PrivilegedLabels, p.PrivilegedLabels)
	k.Taints = util.PickStringMap(k.Taints, p.Taints)
	k.CgroupDriver = util.PickString(k.CgroupDriver, p.CgroupDriver)
	k.NetworkPlugin = util.PickString(k.NetworkPlugin, p.NetworkPlugin, DefaultNetworkPlugin)
	k.SystemReserved = util.PickStringMap(k.SystemReserved, p.SystemReserved)
	k.KubeReserved = util.PickStringMap(k.KubeReserved, p.KubeReserved)
	k.HairpinMode = util.PickString(k.HairpinMode, p.HairpinMode, DefaultHairpinMode)
	k.VolumePluginDir = util.PickString(k.VolumePluginDir, p.VolumePluginDir, defaults.VolumePluginDir)

	if len(k.ExtraMounts) == 0 {
		k.ExtraMounts = p.ExtraMounts
	}

	if len(k.ExtraArgs) == 0 {
		k.ExtraArgs = p.ExtraArgs
	}

	k.Host = host.BuildConfig(k.Host, host.Host{
		SSHConfig: p.SSH,
	})

	p.pkiIntegration()

	p.kubeletPKIIntegration(k)

	if !k.WaitForNodeReady && p.WaitForNodeReady {
		k.WaitForNodeReady = p.WaitForNodeReady
	}
}

// New validates kubelet pool configuration and fills all members with configured values.
func (p *Pool) New() (types.Resource, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("validating pool configuration: %w", err)
	}

	cc := &container.Containers{
		PreviousState: p.State,
		DesiredState:  container.ContainersState{},
	}

	for i := range p.Kubelets {
		k := &p.Kubelets[i]

		p.propagateKubelet(k)

		kubelet, _ := k.New()                                //nolint:errcheck // This is checked in Validate().
		kubeletHcc, _ := kubelet.ToHostConfiguredContainer() //nolint:errcheck // This is checked in Validate().

		cc.DesiredState[strconv.Itoa(i)] = kubeletHcc
	}

	c, _ := cc.New() //nolint:errcheck // This is checked in Validate().

	return &pool{
		containers: c,
	}, nil
}

// Validate validates Pool configuration.
func (p *Pool) Validate() error {
	var errors util.ValidateErrors

	cc := &container.Containers{
		PreviousState: p.State,
		DesiredState:  container.ContainersState{},
	}

	for i := range p.Kubelets {
		// Make a copy of Kubelet struct to avoid modifying original one.
		k := p.Kubelets[i]

		p.propagateKubelet(&k)

		kubelet, err := k.New()
		if err != nil {
			errors = append(errors, fmt.Errorf("creating kubelet object %q: %w", i, err))

			continue
		}

		hcc, err := kubelet.ToHostConfiguredContainer()
		if err != nil {
			errors = append(errors, fmt.Errorf("generating kubelet %q container configuration: %w", i, err))

			continue
		}

		cc.DesiredState[strconv.Itoa(i)] = hcc
	}

	noContainersDefined := len(p.State) == 0 && len(p.Kubelets) == 0
	if noContainersDefined {
		errors = append(errors, fmt.Errorf("at least one kubelet must be defined if state is empty"))
	}

	if _, err := cc.New(); !noContainersDefined && err != nil {
		errors = append(errors, fmt.Errorf("validating containers configuration: %w", err))
	}

	return errors.Return()
}

// FromYaml allows to restore cluster configuration and state from YAML format.
func FromYaml(c []byte) (types.Resource, error) {
	return types.ResourceFromYaml(c, &Pool{})
}

// StateToYaml allows to dump cluster state to YAML, so it can be persisted.
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
