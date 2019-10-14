package kubelet

import (
	"fmt"

	"github.com/invidian/flexkube/pkg/container"
	"github.com/invidian/flexkube/pkg/container/runtime/docker"
	"github.com/invidian/flexkube/pkg/container/types"
	"github.com/invidian/flexkube/pkg/defaults"
	"github.com/invidian/flexkube/pkg/host"
)

// Instance represents single kubelet instance
type Kubelet struct {
	Address             string     `json:"address,omitempty" yaml:"address,omitempty"`
	Image               string     `json:"image,omitempty" yaml:"image,omitempty"`
	Host                *host.Host `json:"host,omitempty" yaml:"host,omitempty"`
	BootstrapKubeconfig string     `json:"bootstrapKubeconfig,omitempty" yaml:"bootstrapKubeconfig,omitempty"`
}

// kubelet is a validated, executable version of Kubelet
type kubelet struct {
	address             string
	image               string
	host                *host.Host
	bootstrapKubeconfig string
}

func (k *Kubelet) New() (*kubelet, error) {
	if err := k.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate kubelet configuration: %w", err)
	}

	nk := &kubelet{
		image:               k.Image,
		address:             k.Address,
		host:                k.Host,
		bootstrapKubeconfig: k.BootstrapKubeconfig,
	}

	if nk.image == "" {
		nk.image = defaults.KubernetesImage
	}

	return nk, nil
}

func (k *Kubelet) Validate() error {
	// TODO better validation should be done here
	if k.BootstrapKubeconfig == "" {
		return fmt.Errorf("bootstrapKubeconfig can't be empty")
	}

	return nil
}

func (k *kubelet) ToHostConfiguredContainer() *container.HostConfiguredContainer {
	configFiles := make(map[string]string)
	// TODO maybe we store that as a struct, and we marshal here to YAML?
	configFiles["/etc/kubernetes/kubelet/kubelet.yaml"] = `apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
staticPodPath: /etc/kubernetes/manifests
rotateCertificates: true
cgroupsPerQOS: false
enforceNodeAllocatable: []
cgroupDriver: systemd
`
	configFiles["/etc/kubernetes/kubelet/bootstrap-kubeconfig"] = k.bootstrapKubeconfig

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: &docker.ClientConfig{},
		},
		Config: types.ContainerConfig{
			// TODO make it configurable?
			Name:  "kubelet",
			Image: k.image,
			// TODO perhaps entrypoint should be a string, not array of strings? we use args for arguments anyway
			Entrypoint: []string{"/kubelet"},
			Ports: []types.PortMap{
				types.PortMap{
					IP:       k.address,
					Protocol: "tcp",
					Port:     10250,
				},
			},
			Privileged: true,
			Mounts: []types.Mount{
				types.Mount{
					Source: "/etc/kubernetes/kubelet/",
					Target: "/etc/kubernetes/",
				},
				types.Mount{
					Source: "/etc/kubernetes/kubelet/manifests/",
					Target: "/etc/kubernetes/manifests",
				},
				types.Mount{
					Source: "/run/docker.sock",
					Target: "/var/run/docker.sock",
				},
				types.Mount{
					Source:      "/var/lib/docker/",
					Target:      "/var/lib/docker",
					Propagation: "rslave",
				},
				types.Mount{
					Source:      "/var/lib/kubelet/",
					Target:      "/var/lib/kubelet",
					Propagation: "shared",
				},
			},
			Args: []string{
				"--config=/etc/kubernetes/kubelet.yaml",
				"--kubeconfig=/etc/kubernetes/kubeconfig",
				"--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubeconfig",
				"--v=2",
				"--network-plugin=kubenet",
			},
		},
	}

	return &container.HostConfiguredContainer{
		Host:        *k.host,
		ConfigFiles: configFiles,
		Container:   c,
	}
}
