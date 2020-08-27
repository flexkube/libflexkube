// Package controlplane allows to create and manage static Kubernetes controlplane running
// in containers.
package controlplane

import (
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/pki"
	"github.com/flexkube/libflexkube/pkg/types"
)

// Common struct contains fields, which are common between all controlplane components.
type Common struct {
	// Image allows to set Docker image with tag, which will be used by all controlplane containers,
	// if they have no image set. If empty, hyperkube image defined in pkg/defaults
	// will be used.
	//
	// Example value: 'k8s.gcr.io/hyperkube:v1.18.3'.
	//
	// This field is optional.
	Image string `json:"image,omitempty"`

	// KubernetesCACertificate stores Kubernetes X.509 CA certificate, PEM encoded.
	//
	// This field is optional.
	KubernetesCACertificate types.Certificate `json:"kubernetesCACertificate,omitempty"`

	// FrontProxyCACertificate stores Kubernetes front proxy X.509 CA certificate, PEM
	// encoded.
	FrontProxyCACertificate types.Certificate `json:"frontProxyCACertificate,omitempty"`
}

// GetImage returns either image defined in common config or Kubernetes default image.
func (co Common) GetImage() string {
	return util.PickString(co.Image, defaults.KubernetesImage)
}

// Controlplane allows creating static Kubernetes controlplane running as containers.
//
// It is usually used to bootstrap self-hosted Kubernetes.
type Controlplane struct {
	// Common stores common fields for all controlplane components. If defined here, the
	// values will be propagated to all 3 components, which allows to de-duplicate parts
	// of the configuration.
	Common *Common `json:"common,omitempty"`

	// SSH stores common SSH configuration for all controlplane components and will be merged
	// with SSH configuration of each component.
	//
	// Usually entire static controlplane runs on a single host, so all values should be defined
	// here.
	//
	// This field is optional.
	SSH *ssh.Config `json:"ssh,omitempty"`

	// APIServerAddress defines Kubernetes API address, which will be used by kube-controller-manager
	// and kube-scheduler to talk to kube-apiserver.
	APIServerAddress string `json:"apiServerAddress,omitempty"`

	// APIServerPort defines Kubernetes API port, which will be used by kube-controller-manager
	// and kube-scheduler to talk to kube-apiserver.
	APIServerPort int `json:"apiServerPort,omitempty"`

	// KubeAPIServer stores kube-apiserver specific configuration.
	KubeAPIServer KubeAPIServer `json:"kubeAPIServer,omitempty"`

	// KubeControllerManager stores kube-controller-manager specific configuration.
	KubeControllerManager KubeControllerManager `json:"kubeControllerManager,omitempty"`

	// KubeScheduler stores kube-scheduler specific configuration.
	KubeScheduler KubeScheduler `json:"kubeScheduler,omitempty"`

	// Destroy controls, if containers should be created or removed. If set to true, all managed
	// containers will be removed.
	Destroy bool `json:"destroy,omitempty"`

	// PKI field allows to use PKI resource for managing all Kubernetes certificates. It will be used for
	// components configuration, if they don't have certificates defined.
	PKI *pki.PKI `json:"pki,omitempty"`

	// State stores state of the created containers. After deployment, it is up to the user to export
	// the state and restore it on consecutive runs.
	State *container.ContainersState `json:"state,omitempty"`
}

// controlplane is executable version of Controlplane, with validated fields and calculated containers.
type controlplane struct {
	containers container.ContainersInterface
}

// propagateKubeconfig merges given client config with values stored in Controlplane.
// Values in given config has priority over ones from the Controlplane.
func (c *Controlplane) propagateKubeconfig(d *client.Config) {
	pkiCA := types.Certificate("")
	if c.PKI != nil && c.PKI.Kubernetes != nil && c.PKI.Kubernetes.CA != nil {
		pkiCA = c.PKI.Kubernetes.CA.X509Certificate
	}

	d.CACertificate = d.CACertificate.Pick(c.Common.KubernetesCACertificate, pkiCA)

	if c.APIServerAddress != "" && c.APIServerPort != 0 {
		d.Server = util.PickString(d.Server, fmt.Sprintf("%s:%d", c.APIServerAddress, c.APIServerPort))
	}
}

// propagateHost merges given host configuration with values stored in Controlplane.
// Values in given host config has priority over ones from the Controlplane.
func (c *Controlplane) propagateHost(h *host.Host) *host.Host {
	if h == nil {
		h = &host.Host{}
	}

	nh := host.BuildConfig(*h, host.Host{
		SSHConfig: c.SSH,
	})

	return &nh
}

// propagateCommon merges given common configuration with values stored in Controlplane.
// Values in given common configuration has priority over ones from the Controlplane.
func (c *Controlplane) propagateCommon(co *Common) {
	if co == nil {
		co = &Common{}
	}

	if c.Common == nil {
		c.Common = &Common{}
	}

	co.Image = util.PickString(co.Image, c.Common.Image)

	var pkiCA types.Certificate
	if c.PKI != nil && c.PKI.Kubernetes != nil && c.PKI.Kubernetes.CA != nil {
		pkiCA = c.PKI.Kubernetes.CA.X509Certificate
	}

	var frontProxyCA types.Certificate
	if c.PKI != nil && c.PKI.Kubernetes != nil && c.PKI.Kubernetes.FrontProxyCA != nil {
		frontProxyCA = c.PKI.Kubernetes.FrontProxyCA.X509Certificate
	}

	co.KubernetesCACertificate = co.KubernetesCACertificate.Pick(c.Common.KubernetesCACertificate, pkiCA)
	co.FrontProxyCACertificate = co.FrontProxyCACertificate.Pick(c.Common.FrontProxyCACertificate, frontProxyCA)
}

// buildKubeScheduler fills KubeSheduler struct with all default values.
func (c *Controlplane) buildKubeScheduler() {
	k := &c.KubeScheduler

	c.propagateKubeconfig(&k.Kubeconfig)

	c.propagateCommon(k.Common)

	// TODO: can be moved to function, which takes Kubeconfig and *pki.Certificate as an input
	if c.PKI != nil && c.PKI.Kubernetes != nil && c.PKI.Kubernetes.KubeSchedulerCertificate != nil {
		k.Kubeconfig.ClientCertificate = k.Kubeconfig.ClientCertificate.Pick(c.PKI.Kubernetes.KubeSchedulerCertificate.X509Certificate)
		k.Kubeconfig.ClientKey = k.Kubeconfig.ClientKey.Pick(c.PKI.Kubernetes.KubeSchedulerCertificate.PrivateKey)
	}

	k.Host = c.propagateHost(k.Host)
}

// buildKubeControllerManager fills KubeControllerManager with all default values.
func (c *Controlplane) buildKubeControllerManager() {
	k := &c.KubeControllerManager

	c.propagateKubeconfig(&k.Kubeconfig)

	c.propagateCommon(k.Common)

	if c.PKI != nil && c.PKI.Kubernetes != nil {
		if c.PKI.Kubernetes.KubeControllerManagerCertificate != nil {
			k.Kubeconfig.ClientCertificate = k.Kubeconfig.ClientCertificate.Pick(c.PKI.Kubernetes.KubeControllerManagerCertificate.X509Certificate)
			k.Kubeconfig.ClientKey = k.Kubeconfig.ClientKey.Pick(c.PKI.Kubernetes.KubeControllerManagerCertificate.PrivateKey)
		}

		if c.PKI.Kubernetes.CA != nil {
			k.KubernetesCAKey = k.KubernetesCAKey.Pick(c.PKI.Kubernetes.CA.PrivateKey)
		}

		if c.PKI.RootCA != nil {
			k.RootCACertificate = k.RootCACertificate.Pick(c.PKI.RootCA.X509Certificate)
		}

		if c.PKI.Kubernetes.ServiceAccountCertificate != nil {
			k.ServiceAccountPrivateKey = k.ServiceAccountPrivateKey.Pick(c.PKI.Kubernetes.ServiceAccountCertificate.PrivateKey)
		}
	}

	k.Host = c.propagateHost(k.Host)

	k.FlexVolumePluginDir = util.PickString(k.FlexVolumePluginDir, defaults.VolumePluginDir)
}

// kubeAPIServerPKIIntegration injects missing certificates and keys from PKI object
// if possible.
func (c *Controlplane) kubeAPIServerPKIIntegration() {
	if c.PKI == nil {
		return
	}

	k := &c.KubeAPIServer

	if p := c.PKI.Etcd; p != nil {
		if p.CA != nil {
			k.EtcdCACertificate = k.EtcdCACertificate.Pick(p.CA.X509Certificate)
		}

		// "root" and "kube-apiserver" are common CNs for etcd client certificate for kube-apiserver.
		for _, cn := range []string{"root", "kube-apiserver"} {
			if c, ok := p.ClientCertificates[cn]; ok {
				k.EtcdClientCertificate = k.EtcdClientCertificate.Pick(c.X509Certificate)
				k.EtcdClientKey = k.EtcdClientKey.Pick(c.PrivateKey)
			}
		}
	}

	if c.PKI.Kubernetes == nil {
		return
	}

	if p := c.PKI.Kubernetes.ServiceAccountCertificate; p != nil {
		k.ServiceAccountPublicKey = util.PickString(k.ServiceAccountPublicKey, p.PublicKey)
	}

	p := c.PKI.Kubernetes.KubeAPIServer
	if p == nil {
		return
	}

	if c := p.ServerCertificate; c != nil {
		k.APIServerCertificate = k.APIServerCertificate.Pick(c.X509Certificate)
		k.APIServerKey = k.APIServerKey.Pick(c.PrivateKey)
	}

	if c := p.FrontProxyClientCertificate; c != nil {
		k.FrontProxyCertificate = k.FrontProxyCertificate.Pick(c.X509Certificate)
		k.FrontProxyKey = k.FrontProxyKey.Pick(c.PrivateKey)
	}

	if c := p.KubeletCertificate; c != nil {
		k.KubeletClientCertificate = k.KubeletClientCertificate.Pick(c.X509Certificate)
		k.KubeletClientKey = k.KubeletClientKey.Pick(c.PrivateKey)
	}
}

// buildKubeAPIServer fills KubeAPIServer with all default values.
func (c *Controlplane) buildKubeAPIServer() {
	k := &c.KubeAPIServer

	if k.BindAddress == "" && c.APIServerAddress != "" {
		k.BindAddress = c.APIServerAddress
	}

	if k.AdvertiseAddress == "" && c.APIServerAddress != "" {
		k.AdvertiseAddress = c.APIServerAddress
	}

	if k.SecurePort == 0 && c.APIServerPort != 0 {
		k.SecurePort = c.APIServerPort
	}

	c.propagateCommon(k.Common)

	c.kubeAPIServerPKIIntegration()

	k.Host = c.propagateHost(k.Host)
}

// New validates Controlplane configuration and fills populates all values provided by the users
// to the structs underneath.
func (c *Controlplane) New() (types.Resource, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate controlplane configuration: %w", err)
	}

	controlplane, cc, _ := c.containersWithState()

	// If shutdown is requested, don't fill DesiredState to remove everything.
	if c.Destroy {
		return controlplane, nil
	}

	// Make sure all values are filled.
	c.buildComponents()

	// Skip error checking, as it's done in Verify().
	kas, _ := c.KubeAPIServer.New()
	kasHcc, _ := kas.ToHostConfiguredContainer()

	kcm, _ := c.KubeControllerManager.New()
	kcmHcc, _ := kcm.ToHostConfiguredContainer()

	ks, _ := c.KubeScheduler.New()
	ksHcc, _ := ks.ToHostConfiguredContainer()

	cc.DesiredState = container.ContainersState{
		"kube-apiserver":          kasHcc,
		"kube-controller-manager": kcmHcc,
		"kube-scheduler":          ksHcc,
	}

	co, _ := cc.New()

	controlplane.containers = co

	return controlplane, nil
}

// buildComponents fills controlplane component structs with default values inherited
// from controlplane struct.
func (c *Controlplane) buildComponents() {
	c.buildKubeAPIServer()
	c.buildKubeControllerManager()
	c.buildKubeScheduler()
}

func (c *Controlplane) containersWithState() (*controlplane, *container.Containers, error) {
	cp := &controlplane{}
	cc := &container.Containers{}

	// If state is empty, just return initialized containers config and controlplane.
	if c.State == nil || len(*c.State) == 0 {
		return cp, cc, nil
	}

	cc.PreviousState = *c.State

	ci, err := cc.New()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create containers state: %w", err)
	}

	cp.containers = ci

	return cp, cc, nil
}

// controlplaneComponentConfiguration is a simple interface wrapping all controlplane
// components.
type controlplaneComponentConfiguration interface {
	New() (container.ResourceInstance, error)
}

// validateControlplaneComponent evaluates configuration of given controlplane component
// and tries to convert it into HostConfiguredContainer to ensure that their configuration
// is correct.
func validateControlplaneComponent(ccc controlplaneComponentConfiguration, name string) (*container.HostConfiguredContainer, error) {
	cc, err := ccc.New()
	if err != nil {
		return nil, fmt.Errorf("failed to verify %q configuration: %w", name, err)
	}

	hcc, err := cc.ToHostConfiguredContainer()
	if err != nil {
		return nil, fmt.Errorf("failed to build %q container configuration: %w", name, err)
	}

	return hcc, nil
}

// Validate validates Controlplane configuration.
func (c *Controlplane) Validate() error {
	c.buildComponents()

	var errors util.ValidateError

	if c.Destroy && (c.State == nil || len(*c.State) == 0) {
		errors = append(errors, fmt.Errorf("can't destroy non-existent controlplane"))
	}

	_, cc, err := c.containersWithState()
	if err != nil {
		errors = append(errors, fmt.Errorf("malformed containers state: %w", err))
	}

	// If we destroy, we only need to validate the state.
	if c.Destroy {
		return errors.Return()
	}

	kasHcc, err := validateControlplaneComponent(&c.KubeAPIServer, "kube-apiserver")
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to verify kube-apiserver configuration: %w", err))
	}

	kcmHcc, err := validateControlplaneComponent(&c.KubeControllerManager, "kube-controller-manager")
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to verify kube-controller-manager: %w", err))
	}

	ksHcc, err := validateControlplaneComponent(&c.KubeScheduler, "kube-scheduler")
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to verify kube-scheduler configuration: %w", err))
	}

	// If there were any errors while creating objects, it's not safe to proceed.
	if len(errors) > 0 {
		return errors.Return()
	}

	cc.DesiredState = container.ContainersState{
		"kube-apiserver":          kasHcc,
		"kube-controller-manager": kcmHcc,
		"kube-scheduler":          ksHcc,
	}

	if _, err = cc.New(); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate containers configuration: %w", err))
	}

	return errors.Return()
}

// FromYaml allows to restore controlplane configuration and state from YAML format.
func FromYaml(c []byte) (types.Resource, error) {
	return types.ResourceFromYaml(c, &Controlplane{})
}

// StateToYaml allows to dump controlplane state to YAML, so it can be restored later.
func (c *controlplane) StateToYaml() ([]byte, error) {
	return yaml.Marshal(Controlplane{State: &c.containers.ToExported().PreviousState})
}

func (c *controlplane) CheckCurrentState() error {
	return c.containers.CheckCurrentState()
}

// Deploy checks the status of the control plane and deploys configuration updates.
func (c *controlplane) Deploy() error {
	return c.containers.Deploy()
}

// Containers implement types.Resource interface.
func (c *controlplane) Containers() container.ContainersInterface {
	return c.containers
}
