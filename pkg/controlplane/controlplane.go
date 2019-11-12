package controlplane

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/direct"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

// Controlplane represents etcd controlplane configuration and state from the user
type Controlplane struct {
	// User-configurable fields
	// They should be defined here if they are used more than once. Things like serviceCIDR, which is only needed in KubeAPIServer,
	// should be defined directly there.
	Image                    string      `json:"image,omitempty" yaml:"image,omitempty"`
	SSH                      *ssh.Config `json:"ssh,omitempty" yaml:"ssh,omitempty"`
	KubernetesCACertificate  string      `json:"kubernetesCACertificate,omitempty" yaml:"kubernetesCACertificate,omitempty"`
	KubernetesCAKey          string      `json:"kubernetesCAKey,omitempty" yaml:"kubernetesCAKey,omitempty"`
	ServiceAccountPublicKey  string      `json:"serviceAccountPublicKey,omitempty" yaml:"serviceAccountPublicKey,omitempty"`
	ServiceAccountPrivateKey string      `json:"serviceAccountPrivateKey,omitempty" yaml:"serviceAccountPrivateKey,omitempty"`
	APIServerCertificate     string      `json:"apiServerCertificate,omitempty" yaml:"apiServerCertificate,omitempty"`
	APIServerKey             string      `json:"apiServerKey,omitempty" yaml:"apiServerKey,omitempty"`
	APIServerAddress         string      `json:"apiServerAddress,omitempty" yaml:"apiServerAddress,omitempty"`
	APIServerPort            int         `json:"apiServerPort,omitempty" yaml:"apiServerPort,omitempty"`
	EtcdServers              []string    `json:"etcdServers,omitempty" yaml:"etcdServers,omitempty"`
	AdminCertificate         string      `json:"adminCertificate,omitempty" yaml:"adminCertificate,omitempty"`
	AdminKey                 string      `json:"adminKey,omitempty" yaml:"adminKey,omitempty"`
	RootCACertificate        string      `json:"rootCACertificate,omitempty" yaml:"rootCACertificate,omitempty"`

	KubeAPIServer         KubeAPIServer         `json:"kubeAPIServer,omitempty" yaml:"kubeAPIServer,omitempty"`
	KubeControllerManager KubeControllerManager `json:"kubeControllerManager,omitempty" yaml:"kubeControllerManager,omitempty"`
	KubeScheduler         KubeScheduler         `json:"kubeScheduler,omitempty" yaml:"kubeScheduler,omitempty"`

	Shutdown bool `json:"shutdown,omitempty" yaml:"shutdown,omitempty"`

	// Serializable fields
	State container.ContainersState `json:"state:omitempty" yaml:"state,omitempty"`
}

// controlplane is executable version of Controlplane, with validated fields and calculated containers
type controlplane struct {
	image                    string
	ssh                      *ssh.Config
	kubernetesCACertificate  string
	kubernetesCAKey          string
	serviceAccountPublicKey  string
	serviceAccountPrivateKey string
	apiServerCertificate     string
	apiServerKey             string
	apiServerAddress         string
	apiServerPort            int
	etcdServers              []string
	adminCertificate         string
	adminKey                 string
	rootCACertificate        string
	shutdown                 bool

	containers container.Containers
}

func (c *Controlplane) buildKubeScheduler() {
	if c.KubeScheduler.Image == "" && c.Image != "" {
		c.KubeScheduler.Image = c.Image
	}
	if c.KubeScheduler.KubernetesCACertificate == "" && c.KubernetesCACertificate != "" {
		c.KubeScheduler.KubernetesCACertificate = c.KubernetesCACertificate
	}
	if c.KubeScheduler.APIServer == "" && c.APIServerAddress != "" {
		c.KubeScheduler.APIServer = c.APIServerAddress
	}
	if c.KubeScheduler.AdminCertificate == "" && c.AdminCertificate != "" {
		c.KubeScheduler.AdminCertificate = c.AdminCertificate
	}
	if c.KubeScheduler.AdminKey == "" && c.AdminKey != "" {
		c.KubeScheduler.AdminKey = c.AdminKey
	}

	// TODO find better way to handle defaults!!!
	if (c.KubeScheduler.Host == nil || (c.KubeScheduler.Host.DirectConfig == nil && c.KubeScheduler.Host.SSHConfig == nil)) && c.SSH == nil {
		c.KubeScheduler.Host = &host.Host{
			DirectConfig: &direct.Config{},
		}
	}
	if c.KubeScheduler.Host == nil {
		c.KubeScheduler.Host = &host.Host{
			SSHConfig: c.SSH,
		}
	}
	if c.KubeScheduler.Host != nil && c.KubeScheduler.Host.SSHConfig != nil && c.KubeScheduler.Host.SSHConfig.PrivateKey == "" && c.SSH != nil && c.SSH.PrivateKey != "" {
		c.KubeScheduler.Host.SSHConfig.PrivateKey = c.SSH.PrivateKey
	}

	if c.KubeScheduler.Host != nil && c.KubeScheduler.Host.SSHConfig != nil && c.KubeScheduler.Host.SSHConfig.User == "" && c.SSH != nil && c.SSH.User != "" {
		c.KubeScheduler.Host.SSHConfig.User = c.SSH.User
	}
	if c.KubeScheduler.Host != nil && c.KubeScheduler.Host.SSHConfig != nil && c.KubeScheduler.Host.SSHConfig.User == "" {
		c.KubeScheduler.Host.SSHConfig.User = "root"
	}

	if c.KubeScheduler.Host != nil && c.KubeScheduler.Host.SSHConfig != nil && c.KubeScheduler.Host.SSHConfig.ConnectionTimeout == "" && c.SSH != nil && c.SSH.ConnectionTimeout != "" {
		c.KubeScheduler.Host.SSHConfig.ConnectionTimeout = c.SSH.ConnectionTimeout
	}
	if c.KubeScheduler.Host != nil && c.KubeScheduler.Host.SSHConfig != nil && c.KubeScheduler.Host.SSHConfig.ConnectionTimeout == "" {
		c.KubeScheduler.Host.SSHConfig.ConnectionTimeout = "30s"
	}

	if c.KubeScheduler.Host != nil && c.KubeScheduler.Host.SSHConfig != nil && c.KubeScheduler.Host.SSHConfig.Port == 0 && c.SSH != nil && c.SSH.Port != 0 {
		c.KubeScheduler.Host.SSHConfig.Port = c.SSH.Port
	}
	if c.KubeScheduler.Host != nil && c.KubeScheduler.Host.SSHConfig != nil && c.KubeScheduler.Host.SSHConfig.Port == 0 {
		c.KubeScheduler.Host.SSHConfig.Port = 22
	}
}

func (c *Controlplane) buildKubeControllerManager() {
	if c.KubeControllerManager.Image == "" && c.Image != "" {
		c.KubeControllerManager.Image = c.Image
	}
	if c.KubeControllerManager.KubernetesCACertificate == "" && c.KubernetesCACertificate != "" {
		c.KubeControllerManager.KubernetesCACertificate = c.KubernetesCACertificate
	}
	if c.KubeControllerManager.KubernetesCAKey == "" && c.KubernetesCAKey != "" {
		c.KubeControllerManager.KubernetesCAKey = c.KubernetesCAKey
	}
	if c.KubeControllerManager.ServiceAccountPrivateKey == "" && c.ServiceAccountPrivateKey != "" {
		c.KubeControllerManager.ServiceAccountPrivateKey = c.ServiceAccountPrivateKey
	}
	if c.KubeControllerManager.AdminCertificate == "" && c.AdminCertificate != "" {
		c.KubeControllerManager.AdminCertificate = c.AdminCertificate
	}
	if c.KubeControllerManager.AdminKey == "" && c.AdminKey != "" {
		c.KubeControllerManager.AdminKey = c.AdminKey
	}

	if c.KubeControllerManager.APIServer == "" && c.APIServerAddress != "" {
		c.KubeControllerManager.APIServer = c.APIServerAddress
	}
	if c.KubeControllerManager.RootCACertificate == "" && c.RootCACertificate != "" {
		c.KubeControllerManager.RootCACertificate = c.RootCACertificate
	}

	// TODO find better way to handle defaults!!!
	if (c.KubeControllerManager.Host == nil || (c.KubeControllerManager.Host.DirectConfig == nil && c.KubeControllerManager.Host.SSHConfig == nil)) && c.SSH == nil {
		c.KubeControllerManager.Host = &host.Host{
			DirectConfig: &direct.Config{},
		}
	}
	if c.KubeControllerManager.Host == nil {
		c.KubeControllerManager.Host = &host.Host{
			SSHConfig: c.SSH,
		}
	}
	if c.KubeControllerManager.Host != nil && c.KubeControllerManager.Host.SSHConfig != nil && c.KubeControllerManager.Host.SSHConfig.PrivateKey == "" && c.SSH != nil && c.SSH.PrivateKey != "" {
		c.KubeControllerManager.Host.SSHConfig.PrivateKey = c.SSH.PrivateKey
	}

	if c.KubeControllerManager.Host != nil && c.KubeControllerManager.Host.SSHConfig != nil && c.KubeControllerManager.Host.SSHConfig.User == "" && c.SSH != nil && c.SSH.User != "" {
		c.KubeControllerManager.Host.SSHConfig.User = c.SSH.User
	}
	if c.KubeControllerManager.Host != nil && c.KubeControllerManager.Host.SSHConfig != nil && c.KubeControllerManager.Host.SSHConfig.User == "" {
		c.KubeControllerManager.Host.SSHConfig.User = "root"
	}

	if c.KubeControllerManager.Host != nil && c.KubeControllerManager.Host.SSHConfig != nil && c.KubeControllerManager.Host.SSHConfig.ConnectionTimeout == "" && c.SSH != nil && c.SSH.ConnectionTimeout != "" {
		c.KubeControllerManager.Host.SSHConfig.ConnectionTimeout = c.SSH.ConnectionTimeout
	}
	if c.KubeControllerManager.Host != nil && c.KubeControllerManager.Host.SSHConfig != nil && c.KubeControllerManager.Host.SSHConfig.ConnectionTimeout == "" {
		c.KubeControllerManager.Host.SSHConfig.ConnectionTimeout = "30s"
	}

	if c.KubeControllerManager.Host != nil && c.KubeControllerManager.Host.SSHConfig != nil && c.KubeControllerManager.Host.SSHConfig.Port == 0 && c.SSH != nil && c.SSH.Port != 0 {
		c.KubeControllerManager.Host.SSHConfig.Port = c.SSH.Port
	}
	if c.KubeControllerManager.Host != nil && c.KubeControllerManager.Host.SSHConfig != nil && c.KubeControllerManager.Host.SSHConfig.Port == 0 {
		c.KubeControllerManager.Host.SSHConfig.Port = 22
	}
}

func (c *Controlplane) buildKubeAPIServer() {
	if c.KubeAPIServer.Image == "" && c.Image != "" {
		c.KubeAPIServer.Image = c.Image
	}
	if c.KubeAPIServer.KubernetesCACertificate == "" && c.KubernetesCACertificate != "" {
		c.KubeAPIServer.KubernetesCACertificate = c.KubernetesCACertificate
	}
	if c.KubeAPIServer.APIServerCertificate == "" && c.APIServerCertificate != "" {
		c.KubeAPIServer.APIServerCertificate = c.APIServerCertificate
	}
	if c.KubeAPIServer.APIServerKey == "" && c.APIServerKey != "" {
		c.KubeAPIServer.APIServerKey = c.APIServerKey
	}
	if c.KubeAPIServer.ServiceAccountPublicKey == "" && c.ServiceAccountPublicKey != "" {
		c.KubeAPIServer.ServiceAccountPublicKey = c.ServiceAccountPublicKey
	}
	if len(c.KubeAPIServer.EtcdServers) == 0 && len(c.EtcdServers) >= 0 {
		c.KubeAPIServer.EtcdServers = c.EtcdServers
	}
	if c.KubeAPIServer.BindAddress == "" && c.APIServerAddress != "" {
		c.KubeAPIServer.BindAddress = c.APIServerAddress
	}
	if c.KubeAPIServer.AdvertiseAddress == "" && c.APIServerAddress != "" {
		c.KubeAPIServer.AdvertiseAddress = c.APIServerAddress
	}
	if c.KubeAPIServer.SecurePort == 0 && c.APIServerPort != 0 {
		c.KubeAPIServer.SecurePort = c.APIServerPort
	}

	// TODO find better way to handle defaults!!!
	if (c.KubeAPIServer.Host == nil || (c.KubeAPIServer.Host.DirectConfig == nil && c.KubeAPIServer.Host.SSHConfig == nil)) && c.SSH == nil {
		c.KubeAPIServer.Host = &host.Host{
			DirectConfig: &direct.Config{},
		}
	}
	if c.KubeAPIServer.Host == nil {
		c.KubeAPIServer.Host = &host.Host{
			SSHConfig: c.SSH,
		}
	}
	if c.KubeAPIServer.Host != nil && c.KubeAPIServer.Host.SSHConfig != nil && c.KubeAPIServer.Host.SSHConfig.PrivateKey == "" && c.SSH != nil && c.SSH.PrivateKey != "" {
		c.KubeAPIServer.Host.SSHConfig.PrivateKey = c.SSH.PrivateKey
	}

	if c.KubeAPIServer.Host != nil && c.KubeAPIServer.Host.SSHConfig != nil && c.KubeAPIServer.Host.SSHConfig.User == "" && c.SSH != nil && c.SSH.User != "" {
		c.KubeAPIServer.Host.SSHConfig.User = c.SSH.User
	}
	if c.KubeAPIServer.Host != nil && c.KubeAPIServer.Host.SSHConfig != nil && c.KubeAPIServer.Host.SSHConfig.User == "" {
		c.KubeAPIServer.Host.SSHConfig.User = "root"
	}

	if c.KubeAPIServer.Host != nil && c.KubeAPIServer.Host.SSHConfig != nil && c.KubeAPIServer.Host.SSHConfig.ConnectionTimeout == "" && c.SSH != nil && c.SSH.ConnectionTimeout != "" {
		c.KubeAPIServer.Host.SSHConfig.ConnectionTimeout = c.SSH.ConnectionTimeout
	}
	if c.KubeAPIServer.Host != nil && c.KubeAPIServer.Host.SSHConfig != nil && c.KubeAPIServer.Host.SSHConfig.ConnectionTimeout == "" {
		c.KubeAPIServer.Host.SSHConfig.ConnectionTimeout = "30s"
	}

	if c.KubeAPIServer.Host != nil && c.KubeAPIServer.Host.SSHConfig != nil && c.KubeAPIServer.Host.SSHConfig.Port == 0 && c.SSH != nil && c.SSH.Port != 0 {
		c.KubeAPIServer.Host.SSHConfig.Port = c.SSH.Port
	}
	if c.KubeAPIServer.Host != nil && c.KubeAPIServer.Host.SSHConfig != nil && c.KubeAPIServer.Host.SSHConfig.Port == 0 {
		c.KubeAPIServer.Host.SSHConfig.Port = 22
	}
}

func (c *Controlplane) New() (*controlplane, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate controlplane configuration: %w", err)
	}

	// If shutdown is requested, only shut down containers.
	if c.Shutdown {
		return &controlplane{
			containers: container.Containers{
				PreviousState: c.State,
				DesiredState:  make(container.ContainersState),
			},
		}, nil
	}

	controlplane := &controlplane{
		image:                    c.Image,
		ssh:                      c.SSH,
		kubernetesCACertificate:  c.KubernetesCACertificate,
		kubernetesCAKey:          c.KubernetesCAKey,
		serviceAccountPublicKey:  c.ServiceAccountPublicKey,
		serviceAccountPrivateKey: c.ServiceAccountPrivateKey,
		apiServerCertificate:     c.APIServerCertificate,
		apiServerKey:             c.APIServerKey,
		apiServerAddress:         c.APIServerAddress,
		apiServerPort:            c.APIServerPort,
		etcdServers:              c.EtcdServers,
		adminCertificate:         c.AdminCertificate,
		adminKey:                 c.AdminKey,
		rootCACertificate:        c.RootCACertificate,
		shutdown:                 c.Shutdown,
		containers: container.Containers{
			PreviousState: c.State,
			DesiredState:  make(container.ContainersState),
		},
	}

	c.buildKubeAPIServer()
	kas, err := c.KubeAPIServer.New()
	if err != nil {
		return nil, fmt.Errorf("that was unexpected: %w", err)
	}
	controlplane.containers.DesiredState["kube-apiserver"] = kas.ToHostConfiguredContainer()

	c.buildKubeControllerManager()
	kcm, err := c.KubeControllerManager.New()
	if err != nil {
		return nil, fmt.Errorf("that was unexpected: %w", err)
	}
	controlplane.containers.DesiredState["kube-controller-manager"] = kcm.ToHostConfiguredContainer()

	c.buildKubeScheduler()
	ks, err := c.KubeScheduler.New()
	if err != nil {
		return nil, fmt.Errorf("that was unexpected: %w", err)
	}
	controlplane.containers.DesiredState["kube-scheduler"] = ks.ToHostConfiguredContainer()

	return controlplane, nil
}

func (c *Controlplane) Validate() error {
	if c.AdminCertificate == "" {
		return fmt.Errorf("AdminCertificate is empty")
	}

	// TODO too many things to validate, let's skip it for now
	return nil
}

// FromYaml allows to restore controlplane state from YAML.
func FromYaml(c []byte) (*controlplane, error) {
	controlplane := &Controlplane{}
	if err := yaml.Unmarshal(c, &controlplane); err != nil {
		return nil, fmt.Errorf("failed to parse input yaml: %w", err)
	}
	cl, err := controlplane.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create controlplane object: %w", err)
	}
	return cl, nil
}

// StateToYaml allows to dump controlplane state to YAML, so it can be restored later.
func (c *controlplane) StateToYaml() ([]byte, error) {
	return yaml.Marshal(Controlplane{State: c.containers.PreviousState})
}

func (c *controlplane) CheckCurrentState() error {
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

func (c *controlplane) Deploy() error {
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
	// Alternatively we can have serializable plan and a knob in execute command to control whether we should
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
