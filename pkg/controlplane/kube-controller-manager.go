package controlplane

import (
	"fmt"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
	"github.com/flexkube/libflexkube/pkg/types"
)

// KubeControllerManager represents kube-controller-manager container configuration.
type KubeControllerManager struct {
	// Common stores common information between all controlplane components.
	Common *Common `json:"common,omitempty"`

	// Host defines on which host kube-controller-manager container should be created.
	Host *host.Host `json:"host,omitempty"`

	// Kubeconfig stores client information used by kube-controller-manager to talk to
	// Kubernetes API.
	Kubeconfig client.Config `json:"kubeconfig"`

	// KubernetesCAKey is a PEM encoded, private key in either PKCS1, PKCS8 or EC format,
	// which was used to sign all Kubernetes certificates. It will be used by
	// kube-controller-manager to sign Kubernetes certificate requests, for example issued by
	// kubelet as part of TLS bootstrapping and rotation process.
	KubernetesCAKey types.PrivateKey `json:"kubernetesCAKey"`

	// ServiceAccountPrivateKey is a PEM encoded, private key in either PKCS1, PKCS8 or EC format,
	// which will be used by to sing service account tokens.
	ServiceAccountPrivateKey types.PrivateKey `json:"serviceAccountPrivateKey"`

	// RootCACertificate is a X.509 CA certificate, PEM encoded, which signed Kubernetes CA
	// certificate. It will be included into service account tokens, so clients like 'curl', can
	// perform full validation of Kubernetes API certificate.
	RootCACertificate types.Certificate `json:"rootCACertificate"`

	// FlexVolumePluginDir is a plugin directory for FlexVolumes, which must be defined for
	// kube-controller-manager, as stated in Flexvolume specification.
	//
	// Example value: '/usr/libexec/kubernetes/kubelet-plugins/volume/exec/'.
	FlexVolumePluginDir string `json:"flexVolumePluginDir"`
}

// kubeControllerManager is a validated version of KubeControllerManager.
type kubeControllerManager struct {
	common                   Common
	host                     host.Host
	kubernetesCAKey          string
	serviceAccountPrivateKey string
	rootCACertificate        string
	kubeconfig               string
	flexVolumePluginDir      string
}

// args returns kube-controller-manager arguments passed to the container.
func (k *kubeControllerManager) args() []string {
	return []string{
		"kube-controller-manager",
		// This makes controller manager use built-in roles, which already has all required
		// roles binded. As kubeconfig file we use should use kube-controller-manager service
		// account, this is required for things to function properly. More info here:
		// https://kubernetes.io/docs/reference/access-authn-authz/rbac/#controller-roles.
		"--use-service-account-credentials",
		// signing-cert and signing-key flags are required for issuing certificates
		// inside cluster. This is for example required for kubelet TLS bootstrapping.
		"--cluster-signing-cert-file=/etc/kubernetes/pki/ca.crt",
		"--cluster-signing-key-file=/etc/kubernetes/pki/ca.key",
		// Specifies private RSA key which will be used for signing service account tokens,
		// as one of kube-controller-manager roles is to create tokens for each service account.
		//
		// Kubernetes API server has private key configured for verification.
		"--service-account-private-key-file=/etc/kubernetes/pki/service-account.key",
		// Specifies which certificate will be included in service account secrets, which will be used,
		// to establish trust to API server. This should point to the file containing both Kubernetes CA certificate,
		// and root CA certificate, as otherwise clients won't trust kube-apiserver service certificate.
		"--root-ca-file=/etc/kubernetes/pki/root.crt",
		// This kubeconfig file will be used for talking to API server.
		"--kubeconfig=/etc/kubernetes/kubeconfig",
		// Those additional kubeconfig files are suppose to be used with delegated kube-apiserver,
		// so scenarios, where there is more than one kube-apiserver and they differ in privilege level.
		// However, not specifying them results in ugly log messages, so we just specify them to create less
		// environmental noise.
		"--authentication-kubeconfig=/etc/kubernetes/kubeconfig",
		"--authorization-kubeconfig=/etc/kubernetes/kubeconfig",
		// From k8s 1.17.x, without specifying those flags, there are some warning log messages printed.
		"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
		"--client-ca-file=/etc/kubernetes/pki/ca.crt",
		fmt.Sprintf("--flex-volume-plugin-dir=%s", k.flexVolumePluginDir),
	}
}

// ToHostConfiguredContainer takes configured parameters and returns generic HostConfiguredContainer.
//
// TODO refactor this method, to have a generic method, which takes host as an argument and returns you
// a HostConfiguredContainer with hyperkube image configured, initialized configFiles map etc.
func (k *kubeControllerManager) ToHostConfiguredContainer() (*container.HostConfiguredContainer, error) {
	configFiles := map[string]string{}
	// TODO put all those path in a single place. Perhaps make them configurable with defaults too
	configFiles["/etc/kubernetes/kube-controller-manager/kubeconfig"] = k.kubeconfig
	configFiles["/etc/kubernetes/kube-controller-manager/pki/service-account.key"] = k.serviceAccountPrivateKey
	configFiles["/etc/kubernetes/kube-controller-manager/pki/ca.crt"] = string(k.common.KubernetesCACertificate)
	configFiles["/etc/kubernetes/kube-controller-manager/pki/ca.key"] = k.kubernetesCAKey

	caBundle := fmt.Sprintf("%s%s", k.rootCACertificate, string(k.common.KubernetesCACertificate))
	configFiles["/etc/kubernetes/kube-controller-manager/pki/root.crt"] = caBundle

	frontProxyCA := string(k.common.FrontProxyCACertificate)

	configFiles["/etc/kubernetes/kube-controller-manager/pki/front-proxy-ca.crt"] = frontProxyCA

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: docker.DefaultConfig(),
		},
		Config: containertypes.ContainerConfig{
			Name:  "kube-controller-manager",
			Image: util.PickString(k.common.Image, defaults.KubeControllerManagerImage),
			Mounts: []containertypes.Mount{
				{
					Source: "/etc/kubernetes/kube-controller-manager/",
					Target: "/etc/kubernetes",
				},
			},
			Args: k.args(),
		},
	}

	return &container.HostConfiguredContainer{
		Host:        k.host,
		ConfigFiles: configFiles,
		Container:   c,
	}, nil
}

// New validates KubeControllerManager and returns usable kubeControllerManager.
func (k *KubeControllerManager) New() (container.ResourceInstance, error) {
	if k.Common == nil {
		k.Common = &Common{}
	}

	if k.Host == nil {
		k.Host = &host.Host{}
	}

	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("validating Kubernetes Controller Manager configuration: %w", err)
	}

	kubeconfig, _ := k.Kubeconfig.ToYAMLString() //nolint:errcheck // We check it in Validate().

	nk := &kubeControllerManager{
		common:                   *k.Common,
		host:                     *k.Host,
		kubernetesCAKey:          string(k.KubernetesCAKey),
		serviceAccountPrivateKey: string(k.ServiceAccountPrivateKey),
		rootCACertificate:        string(k.RootCACertificate),
		kubeconfig:               kubeconfig,
		flexVolumePluginDir:      k.FlexVolumePluginDir,
	}

	return nk, nil
}

// Validate validates KubeControllerManager configuration.
func (k *KubeControllerManager) Validate() error {
	v := validator{
		Common:     k.Common,
		Host:       k.Host,
		Kubeconfig: k.Kubeconfig,
		YAML:       k,
	}

	return v.validate(true)
}
