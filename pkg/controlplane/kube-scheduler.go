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

// KubeScheduler represents kube-scheduler configuration data
type KubeScheduler struct {
	Image                   string     `json:"image,omitempty" yaml:"image,omitempty"`
	Host                    *host.Host `json:"host,omitempty" yaml:"host,omitempty"`
	KubernetesCACertificate string     `json:"kubernetesCACertificate,omitempty" yaml:"kubernetesCACertificate,omitempty"`
	APIServer               string     `json:"apiServer,omitempty" yaml:"apiServer,omitempty"`
	// TODO don't take the admin key, use dedicated certificate for static controller manager,
	// which will have a group + create a binding to system:kube-controller-manager clusterRole
	// as done in self-hosted chart.
	// TODO since we have access to CA cert and key, we could generate certificate ourselves here
	ClientCertificate       string `json:"clientCertificate,omitempty" yaml:"clientCertificate,omitempty"`
	ClientKey               string `json:"clientKey,omitempty" yaml:"clientKey,omitempty"`
	FrontProxyCACertificate string `json:"frontProxyCACertificate,omitempty" yaml:"frontProxyCACertificate,omitempty"`
}

// kubeScheduler is validated and usable version of KubeScheduler
type kubeScheduler struct {
	image                   string
	host                    host.Host
	kubernetesCACertificate string
	apiServer               string
	clientCertificate       string
	clientKey               string
	frontProxyCACertificate string
}

// ToHostConfiguredContainer converts kubeScheduler into generic container struct
func (k *kubeScheduler) ToHostConfiguredContainer() *container.HostConfiguredContainer {
	configFiles := make(map[string]string)
	// TODO put all those path in a single place. Perhaps make them configurable with defaults too
	configFiles["/etc/kubernetes/kube-scheduler/kubeconfig"] = k.toKubeconfig()
	configFiles["/etc/kubernetes/kube-scheduler/pki/ca.crt"] = k.kubernetesCACertificate
	configFiles["/etc/kubernetes/kube-scheduler/pki/front-proxy-ca.crt"] = k.frontProxyCACertificate
	configFiles["/etc/kubernetes/kube-scheduler/kube-scheduler.yaml"] = `apiVersion: kubescheduler.config.k8s.io/v1alpha1
kind: KubeSchedulerConfiguration
clientConnection:
  kubeconfig: /etc/kubernetes/kubeconfig
`

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			Name:  "kube-scheduler",
			Image: k.image,
			Mounts: []types.Mount{
				{
					Source: "/etc/kubernetes/kube-scheduler/",
					Target: "/etc/kubernetes",
				},
			},
			Args: []string{
				"kube-scheduler",
				// Load configuration from the config file.
				"--config=/etc/kubernetes/kube-scheduler.yaml",
				// Those additional kubeconfig files are suppose to be used with delegated kube-apiserver,
				// so scenarios, where there is more than one kube-apiserver and they differ in privilege level.
				// However, not specifying them results in ugly log messages, so we just specify them to create less
				// environmental noise.
				"--authentication-kubeconfig=/etc/kubernetes/kubeconfig",
				"--authorization-kubeconfig=/etc/kubernetes/kubeconfig",
				// From k8s 1.17.x, without specifying those flags, there are some warning log messages printed.
				"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
				"--client-ca-file=/etc/kubernetes/pki/ca.crt",
			},
		},
	}

	return &container.HostConfiguredContainer{
		Host:        k.host,
		ConfigFiles: configFiles,
		Container:   c,
	}
}

// New validates KubeScheduler struct and returns it's usable version
func (k *KubeScheduler) New() (*kubeScheduler, error) {
	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate Kubernetes Scheduler configuration: %w", err)
	}

	nk := &kubeScheduler{
		image:                   k.Image,
		host:                    *k.Host,
		kubernetesCACertificate: k.KubernetesCACertificate,
		apiServer:               k.APIServer,
		clientCertificate:       k.ClientCertificate,
		clientKey:               k.ClientKey,
		frontProxyCACertificate: k.FrontProxyCACertificate,
	}

	// The only optional parameter
	if nk.image == "" {
		nk.image = defaults.KubernetesImage
	}

	return nk, nil
}

// Validate valides kube-scheduler configuration
//
// TODO add validation of certificates if specified
func (k *KubeScheduler) Validate() error {
	if k.KubernetesCACertificate == "" {
		return fmt.Errorf("field kubernetesCACertificate is empty")
	}

	if k.APIServer == "" {
		return fmt.Errorf("field apiServer is empty")
	}

	if k.ClientCertificate == "" {
		return fmt.Errorf("field clientCertificate is empty")
	}

	if k.ClientKey == "" {
		return fmt.Errorf("field clientKey is empty")
	}

	if k.FrontProxyCACertificate == "" {
		return fmt.Errorf("field frontProxyCACertificate is empty")
	}

	return nil
}

// toKubeconfig takes given configuration and returns kubeconfig file content for
// kube-scheduler in YAML format
//
// TODO this is quite generic, refactor it
func (k *kubeScheduler) toKubeconfig() string {
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
`, k.apiServer, base64.StdEncoding.EncodeToString([]byte(k.kubernetesCACertificate)), base64.StdEncoding.EncodeToString([]byte(k.clientCertificate)), base64.StdEncoding.EncodeToString([]byte(k.clientKey)))
}
