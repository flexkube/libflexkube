package controlplane

import (
	"encoding/base64"
	"fmt"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
)

// KubeControllerManager represents kube-controller-manager container configuration
type KubeControllerManager struct {
	Image                    string     `json:"image,omitempty" yaml:"image,omitempty"`
	Host                     *host.Host `json:"host,omitempty" yaml:"host,omitempty"`
	KubernetesCACertificate  string     `json:"kubernetesCACertificate,omitempty" yaml:"kubernetesCACertificate,omitempty"`
	KubernetesCAKey          string     `json:"kubernetesCAKey,omitempty" yaml:"kubernetesCAKey,omitempty"`
	ServiceAccountPrivateKey string     `json:"serviceAccountPrivateKey,omitempty" yaml:"serviceAccountPrivateKey,omitempty"`
	APIServer                string     `json:"apiServer,omitempty" yaml:"apiServer,omitempty"`
	// TODO don't take the admin key, use dedicated certificate for static controller manager,
	// which will have a group + create a binding to system:kube-controller-manager clusterRole
	// as done in self-hosted chart.
	// TODO since we have access to CA cert and key, we could generate certificate ourselves here
	AdminCertificate  string `json:"adminCertificate,omitempty" yaml:"adminCertificate,omitempty"`
	AdminKey          string `json:"adminKey,omitempty" yaml:"adminKey,omitempty"`
	RootCACertificate string `josn:"rootCACertificate,omitempty" yaml:"rootCACertificate,omitempty"`
}

// kubeControllerManager is a validated version of KubeControllerManager
type kubeControllerManager struct {
	image                    string
	host                     host.Host
	kubernetesCACertificate  string
	kubernetesCAKey          string
	serviceAccountPrivateKey string
	apiServer                string
	adminCertificate         string
	adminKey                 string
	rootCACertificate        string
}

// ToHostConfiguredContainer takes configured parameters and returns generic HostCOnfiguredContainer
//
// TODO refactor this method, to have a generic method, which takes host as an argument and returns you
// a HostConfiguredContainer with hyperkube image configured, initialized configFiles map etc.
func (k *kubeControllerManager) ToHostConfiguredContainer() *container.HostConfiguredContainer {
	configFiles := make(map[string]string)
	// TODO put all those path in a single place. Perhaps make them configurable with defaults too
	configFiles["/etc/kubernetes/kube-controller-manager/kubeconfig"] = k.toKubeconfig()
	configFiles["/etc/kubernetes/kube-controller-manager/pki/service-account.key"] = k.serviceAccountPrivateKey
	configFiles["/etc/kubernetes/kube-controller-manager/pki/ca.crt"] = k.kubernetesCACertificate
	configFiles["/etc/kubernetes/kube-controller-manager/pki/ca.key"] = k.kubernetesCAKey
	configFiles["/etc/kubernetes/kube-controller-manager/pki/root.crt"] = fmt.Sprintf("%s%s", k.rootCACertificate, k.kubernetesCACertificate)

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:       "kube-controller-manager",
			Image:      k.image,
			Entrypoint: []string{"/hyperkube"},
			Mounts: []types.Mount{
				{
					Source: "/etc/kubernetes/kube-controller-manager/kubeconfig",
					Target: "/etc/kubernetes/kubeconfig",
				},
				{
					Source: "/etc/kubernetes/kube-controller-manager/pki",
					Target: "/etc/kubernetes/pki",
				},
			},
			Args: []string{
				"kube-controller-manager",
				// This kubeconfig file will be used for talking to API server
				"--kubeconfig=/etc/kubernetes/kubeconfig",
				// signing-cert and signing-key flags are required for issuing certificates
				// inside cluster. This is for example required for kubelet TLS bootstrapping.
				"--cluster-signing-cert-file=/etc/kubernetes/pki/ca.crt",
				"--cluster-signing-key-file=/etc/kubernetes/pki/ca.key",
				// Specifies private RSA key which will be used for signing service account tokens,
				// as one of kube-controller-manager roles is to create tokens for each service account.
				//
				// Kubernetes API server has private key configured for verification.
				"--service-account-private-key-file=/etc/kubernetes/pki/service-account.key",
				"--root-ca-file=/etc/kubernetes/pki/root.crt",
			},
		},
	}

	return &container.HostConfiguredContainer{
		Host:        k.host,
		ConfigFiles: configFiles,
		Container:   c,
	}
}

// New validates KubeControllerManager and returns usable kubeControllerManager
func (k *KubeControllerManager) New() (*kubeControllerManager, error) {
	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate Kubernetes Controller Manager configuration: %w", err)
	}

	nk := &kubeControllerManager{
		image:                    k.Image,
		host:                     *k.Host,
		kubernetesCACertificate:  k.KubernetesCACertificate,
		kubernetesCAKey:          k.KubernetesCAKey,
		serviceAccountPrivateKey: k.ServiceAccountPrivateKey,
		apiServer:                k.APIServer,
		adminCertificate:         k.AdminCertificate,
		adminKey:                 k.AdminKey,
		rootCACertificate:        k.RootCACertificate,
	}

	// The only optional parameter
	if nk.image == "" {
		nk.image = defaults.KubernetesImage
	}

	return nk, nil
}

// Validate validates KubeControllerManager configuration
//
// TODO add validation of certificates if specified
func (k *KubeControllerManager) Validate() error {
	if k.KubernetesCACertificate == "" {
		return fmt.Errorf("KubernetesCACertificate is empty")
	}
	if k.KubernetesCAKey == "" {
		return fmt.Errorf("KubernetesCAKey is empty")
	}
	if k.ServiceAccountPrivateKey == "" {
		return fmt.Errorf("ServiceAccountPrivateKey is empty")
	}
	if k.APIServer == "" {
		return fmt.Errorf("APIServer is empty")
	}
	if k.AdminCertificate == "" {
		return fmt.Errorf("AdminCertificate is empty")
	}
	if k.AdminKey == "" {
		return fmt.Errorf("AdminKey is empty")
	}
	if k.RootCACertificate == "" {
		return fmt.Errorf("rootCACertificate is empty")
	}

	return nil
}

// toKubeconfig converts given configuration to kubeconfig format as YAML text
func (k *kubeControllerManager) toKubeconfig() string {
	return fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: static
  cluster:
    server: https://%s:6443
    certificate-authority-data: %s
users:
- name: static
  user:
    client-certificate-data: %s
    client-key-data: %s
current-context: static
contexts:
- name: static
  context:
    cluster: static
    user: static
`, k.apiServer, base64.StdEncoding.EncodeToString([]byte(k.kubernetesCACertificate)), base64.StdEncoding.EncodeToString([]byte(k.adminCertificate)), base64.StdEncoding.EncodeToString([]byte(k.adminKey)))
}
