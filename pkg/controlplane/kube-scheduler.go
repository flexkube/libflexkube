package controlplane

import (
	"fmt"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
)

// KubeScheduler represents kube-scheduler configuration data
type KubeScheduler struct {
	Common     Common        `json:"common" yaml:"common"`
	Host       host.Host     `json:"host" yaml:"host"`
	Kubeconfig client.Config `json:"kubeconfig" yaml:"kubeconfig"`
}

// kubeScheduler is validated and usable version of KubeScheduler
type kubeScheduler struct {
	common     Common
	host       host.Host
	kubeconfig string
}

// ToHostConfiguredContainer converts kubeScheduler into generic container struct
func (k *kubeScheduler) ToHostConfiguredContainer() *container.HostConfiguredContainer {
	configFiles := make(map[string]string)
	// TODO put all those path in a single place. Perhaps make them configurable with defaults too
	configFiles["/etc/kubernetes/kube-scheduler/kubeconfig"] = k.kubeconfig
	configFiles["/etc/kubernetes/kube-scheduler/pki/ca.crt"] = string(k.common.KubernetesCACertificate)
	configFiles["/etc/kubernetes/kube-scheduler/pki/front-proxy-ca.crt"] = string(k.common.FrontProxyCACertificate)
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
		Config: containertypes.ContainerConfig{
			Name:  "kube-scheduler",
			Image: k.common.GetImage(),
			Mounts: []containertypes.Mount{
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
func (k *KubeScheduler) New() (container.ResourceInstance, error) {
	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate Kubernetes Scheduler configuration: %w", err)
	}

	// It's fine to skip the error, Validate() will handle it.
	kubeconfig, _ := k.Kubeconfig.ToYAMLString()

	return &kubeScheduler{
		common:     k.Common,
		host:       k.Host,
		kubeconfig: kubeconfig,
	}, nil
}

// Validate valides kube-scheduler configuration.
func (k *KubeScheduler) Validate() error {
	if _, err := k.Kubeconfig.ToYAMLString(); err != nil {
		return fmt.Errorf("invalid kubeconfig: %w", err)
	}

	if err := k.Host.Validate(); err != nil {
		return fmt.Errorf("host config validation failed: %w", err)
	}

	return nil
}
