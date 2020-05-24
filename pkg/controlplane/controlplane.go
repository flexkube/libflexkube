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
	Image                   string            `json:"image,omitempty"`
	KubernetesCACertificate types.Certificate `json:"kubernetesCACertificate,omitempty"`
	FrontProxyCACertificate types.Certificate `json:"frontProxyCACertificate,omitempty"`
}

// GetImage returns either image defined in common config or Kubernetes default image.
func (co Common) GetImage() string {
	return util.PickString(co.Image, defaults.KubernetesImage)
}

// Controlplane represents kubernetes controlplane configuration and state from the user.
type Controlplane struct {
	// User-configurable fields.
	// They should be defined here if they are used more than once. Things like serviceCIDR, which is only needed in KubeAPIServer,
	// should be defined directly there.
	Common                *Common               `json:"common,omitempty"`
	SSH                   *ssh.Config           `json:"ssh,omitempty"`
	APIServerAddress      string                `json:"apiServerAddress,omitempty"`
	APIServerPort         int                   `json:"apiServerPort,omitempty"`
	KubeAPIServer         KubeAPIServer         `json:"kubeAPIServer,omitempty"`
	KubeControllerManager KubeControllerManager `json:"kubeControllerManager,omitempty"`
	KubeScheduler         KubeScheduler         `json:"kubeScheduler,omitempty"`
	Hosts                 []host.Host           `json:"hosts,omitempty"`
	Destroy               bool                  `json:"destroy,omitempty"`
	PKI                   *pki.PKI              `json:"pki,omitempty"`

	// Serializable fields.
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

	controlplane := &controlplane{}

	cc := &container.Containers{}

	if c.State != nil {
		cc.PreviousState = *c.State

		co, _ := cc.New()

		controlplane.containers = co
	}

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

// Validate validates Controlplane configuration.
func (c *Controlplane) Validate() error {
	c.buildComponents()

	var errors util.ValidateError

	if c.Destroy && (c.State == nil || len(*c.State) == 0) {
		errors = append(errors, fmt.Errorf("can't destroy non-existent controlplane"))
	}

	cc := &container.Containers{}

	if c.State != nil && len(*c.State) > 0 {
		cc.PreviousState = *c.State

		if _, err := cc.New(); err != nil {
			errors = append(errors, fmt.Errorf("unable to create containers state: %w", err))
		}
	}

	kas, err := c.KubeAPIServer.New()
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to verify kube-apiserver configuration: %w", err))
	}

	kcm, err := c.KubeControllerManager.New()
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to verify kube-controller-manager: %w", err))
	}

	ks, err := c.KubeScheduler.New()
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to verify kube-scheduler configuration: %w", err))
	}

	// If there were any errors while creating objects, it's not safe to proceed.
	if len(errors) > 0 {
		return errors.Return()
	}

	kasHcc, err := kas.ToHostConfiguredContainer()
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to build kube-apiserver container configuration: %w", err))
	}

	kcmHcc, err := kcm.ToHostConfiguredContainer()
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to build kube-controller-manager container configuration: %w", err))
	}

	ksHcc, err := ks.ToHostConfiguredContainer()
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to build kube-scheduler container configuration: %w", err))
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

// FromYaml allows to restore controlplane state from YAML.
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
