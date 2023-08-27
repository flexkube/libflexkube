package controlplane

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbaseconfig "k8s.io/component-base/config/v1alpha1"
	kubeschedulerconfig "k8s.io/kube-scheduler/config/v1"
	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	containertypes "github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/kubernetes/client"
)

// KubeScheduler represents kube-scheduler configuration data.
type KubeScheduler struct {
	// Common stores common information between all controlplane components.
	Common *Common `json:"common,omitempty"`

	// Host defines on which host kube-scheduler container should be created.
	Host *host.Host `json:"host,omitempty"`

	// Kubeconfig stores client information used by kube-scheduler to talk to
	// Kubernetes API.
	Kubeconfig client.Config `json:"kubeconfig"`
}

// kubeScheduler is validated and usable version of KubeScheduler.
type kubeScheduler struct {
	common     Common
	host       host.Host
	kubeconfig string
}

// ToHostConfiguredContainer converts kubeScheduler into generic container struct.
func (k *kubeScheduler) ToHostConfiguredContainer() (*container.HostConfiguredContainer, error) {
	configFiles := map[string]string{}
	// TODO put all those path in a single place. Perhaps make them configurable with defaults too
	configFiles["/etc/kubernetes/kube-scheduler/kubeconfig"] = k.kubeconfig
	configFiles["/etc/kubernetes/kube-scheduler/pki/ca.crt"] = string(k.common.KubernetesCACertificate)
	configFiles["/etc/kubernetes/kube-scheduler/pki/front-proxy-ca.crt"] = string(k.common.FrontProxyCACertificate)

	config := &kubeschedulerconfig.KubeSchedulerConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeSchedulerConfiguration",
			APIVersion: kubeschedulerconfig.SchemeGroupVersion.String(),
		},
		ClientConnection: componentbaseconfig.ClientConnectionConfiguration{
			Kubeconfig: "/etc/kubernetes/kubeconfig",
		},
	}

	configRaw, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshaling configuration: %w", err)
	}

	configFiles["/etc/kubernetes/kube-scheduler/kube-scheduler.yaml"] = string(configRaw)

	containerConfig := container.Container{
		// TODO: This is weird. This sets docker as default runtime config.
		Runtime: container.RuntimeConfig{
			Docker: docker.DefaultConfig(),
		},
		Config: containertypes.ContainerConfig{
			Name:  "kube-scheduler",
			Image: util.PickString(k.common.Image, defaults.KubeSchedulerImage),
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
		Container:   containerConfig,
	}, nil
}

// New validates KubeScheduler struct and returns it's usable version.
func (k *KubeScheduler) New() (container.ResourceInstance, error) {
	if k.Common == nil {
		k.Common = &Common{}
	}

	if k.Host == nil {
		k.Host = &host.Host{}
	}

	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("validating Kubernetes Scheduler configuration: %w", err)
	}

	kubeconfig, _ := k.Kubeconfig.ToYAMLString() //nolint:errcheck // We check it in Validate().

	return &kubeScheduler{
		common:     *k.Common,
		host:       *k.Host,
		kubeconfig: kubeconfig,
	}, nil
}

// Validate validates kube-scheduler configuration.
func (k *KubeScheduler) Validate() error {
	schedulerValidator := validator{
		Common:     k.Common,
		Host:       k.Host,
		Kubeconfig: k.Kubeconfig,
		YAML:       k,
	}

	return schedulerValidator.validate(true)
}
